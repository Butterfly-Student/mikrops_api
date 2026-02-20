# Hotspot Management Library

Reusable MikroTik Hotspot Management Library for Go. This library follows the Mikhmon v3 API patterns for managing hotspot users, profiles, vouchers, sessions, and sales records.

## Features

- **Profile Management**: Create, update, delete, and list hotspot user profiles
- **User Management**: Full CRUD operations with filtering capabilities
- **Voucher Generation**: Batch generation with customizable random credentials
- **Active Sessions Monitoring**: Real-time session tracking and management
- **Sales Recording**: Store sales records as RouterOS system scripts
- **Expiry Management**: Automated scheduler-based expiry monitoring

## Architecture

The library follows a clean architecture with:
- **Interface-based design** for easy testing and mocking
- **Context support** for cancellation and timeouts
- **Structured error handling** with wrapped errors
- **Internal package** for parser/builder utilities (no import cycles)

### RouterOS Data Storage

All data is stored directly in MikroTik RouterOS:

| Entity | RouterOS Path |
|--------|---------------|
| Profiles | `/ip/hotspot/user/profile` |
| Users | `/ip/hotspot/user` |
| Sessions | `/ip/hotspot/active` |
| Sales | `/system/script` (special naming format) |
| Schedulers | `/system/scheduler` |

No additional database required.

## Usage

### Basic Setup

```go
import (
    "context"
    "go-template/pkg/hotspot"
    routeros "github.com/go-routeros/routeros/v3"
)

// Create MikroTik connection
mikrotikClient, err := routeros.Dial("192.168.1.1:8728", "admin", "password")
if err != nil {
    log.Fatal(err)
}
defer mikrotikClient.Close()

// Create hotspot client
client := hotspot.NewClient(1, mikrotikClient)
```

### Profile Management

```go
ctx := context.Background()

// Create a new profile
profile := &hotspot.Profile{
    Name:             "prepaid-daily",
    SharedUsers:      1,
    RateLimit:        "1M/2M",
    Validity:         "1d",
    Price:            5000,
    SellingPrice:     7000,
    ExpiryMode:       hotspot.ExpiryModeRemove,
    LockUser:         "yes",
    KeepaliveTimeout: "00:05:00",
}
err := client.CreateProfile(ctx, profile)

// Get all profiles
profiles, err := client.GetAllProfiles(ctx)

// Update profile
updates := &hotspot.ProfileUpdate{
    RateLimit: stringPtr("2M/4M"),
    Price:     float64Ptr(6000),
}
err := client.UpdateProfile(ctx, "prepaid-daily", updates)
```

### User Management

```go
ctx := context.Background()

// Create single user
user := &hotspot.User{
    Name:            "test-user",
    Password:        "pass123",
    Profile:         "prepaid-daily",
    Comment:         "Jan/20/2026 vc-TEST",
    LimitUptime:     3600,  // 1 hour in seconds
    LimitBytesTotal:  100000000, // 100MB
    Disabled:        false,
    Server:          "all",
}
err := client.CreateUser(ctx, user)

// Get user
user, err := client.GetUser(ctx, "test-user")

// Get all users with filter (server-side: Profile, Disabled; client-side: Limit/Offset pagination)
users, err := client.GetAllUsers(ctx, &hotspot.UserFilter{
    Profile:  "prepaid-daily",
    Disabled: boolPtr(false),
    Limit:    100,
})

// Search users by partial comment substring (client-side filter)
users, err := client.GetUsersByComment(ctx, "vc-CAFE")

// Update user
err := client.UpdateUser(ctx, "test-user", &hotspot.UserUpdate{
    Disabled: boolPtr(true),
})

// Delete user
err := client.DeleteUser(ctx, "test-user")
```

### Voucher Generation

```go
ctx := context.Background()

// Generate voucher mode (username == password)
gen := &hotspot.VoucherGenerator{
    Profile:         "prepaid-daily",
    Prefix:          "CAFE",   // optional; leave empty for no prefix
    Charset:         hotspot.DefaultCharset,
    LengthUsername:  8,
    LengthPassword:  8,
    Quantity:        50,
    TimeLimit:       86400,    // 1 day in seconds (0 = unlimited)
    DataLimit:       0,        // unlimited
    Validity:        "1d",     // expiry: "1d", "2d", "1w", "2w", "1y", "2h", "30m"
}
result, err := client.GenerateVouchers(ctx, gen)

fmt.Printf("Generated %d vouchers, %d failed\n", result.Success, result.Failed)
for _, voucher := range result.Vouchers {
    fmt.Printf("Username: %s, Password: %s\n", voucher.Name, voucher.Password)
}
```

### Session Monitoring

```go
ctx := context.Background()

// Get all active sessions
sessions, err := client.GetActiveSessions(ctx)

// Get sessions by server
sessions, err := client.GetSessionsByServer(ctx, "hs1")

// Get specific user session
session, err := client.GetSessionByUsername(ctx, "test-user")

// Disconnect user
err := client.DisconnectUser(ctx, "test-user")

// Get session statistics
stats, err := client.GetSessionStats(ctx)
fmt.Printf("Active users: %d\n", stats.ActiveUsers)
```

### Sales Recording

```go
ctx := context.Background()

// Record a sale
sale := &hotspot.Sale{
    Username: "CAFE-ABCD1234",
    Price:    7000,
    Address:  "192.168.1.100",
    Mac:      "AA:BB:CC:DD:EE:FF",
    Validity: "1d",
}
err := client.RecordSale(ctx, sale)

// Get sales by date range
sales, err := client.GetSalesByDateRange(ctx, "jan/01/2026", "jan/31/2026")

// Get total revenue
revenue, err := client.GetTotalRevenue(ctx, "jan/01/2026", "jan/31/2026")
fmt.Printf("Total revenue: %.2f\n", revenue)
```

### Expiry Management

```go
ctx := context.Background()

// Create scheduler for automated expiry checking
err := client.CreateExpiryScheduler(ctx, "prepaid-daily")

// Remove expired users manually
removed, err := client.RemoveExpiredUsers(ctx, "prepaid-daily")
fmt.Printf("Removed %d expired users\n", removed)

// Remove unused vouchers
removed, err := client.RemoveUnusedVouchers(ctx, "prepaid-daily")
```

## Configuration Formats

### Profile On-Login Script

Profile settings are stored in the `on-login` script:

```
:local expmode "rem";:local price "5000.00";:local validity "1d";:local selling "7000.00";:local lock "yes";
```

### Sales Record Script Name

Sales are stored as RouterOS scripts with special naming (owner = `hotspot-sales`).
All 7 fields are always present; optional fields may be empty strings:

```
date-|-time-|-username-|-price-|-address-|-mac-|-validity
```

Example: `Jan/20/2026-|-16:05:11-|-CAFE-ABCD-|-7000.00-|-192.168.1.100-|-AA:BB:CC:DD:EE:FF-|-1d`

Example with empty optional fields: `Jan/20/2026-|-10:00:00-|-USER-ABC-|-5000.00-|-  -|-  -|-2d`

### User Comment Format

User comments store expiry date and user mode:

```
DATE MODE-PREFIX
```

Examples:
- `Jan/20/2026 vc-` - Voucher mode (username == password)
- `Jan/20/2026 up-PREFIX` - User-Password mode (username != password)

## Expiry Modes

| Mode | Description |
|------|-------------|
| `rem` | Remove expired users automatically |
| `ntf` | Notify expired users only |
| `remc` | Remove expired users and copy to script |
| `ntfc` | Notify expired users and copy to script |

## Constants

```go
// Character sets
const (
    DefaultCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"  // No confusing chars
    NumericCharset = "0123456789"
    Alphanumeric   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Length constraints
const (
    MinUsername = 4
    MaxUsername = 12
    MinPassword = 4
    MaxPassword = 12
)

// Default values
const (
    DefaultServer            = "all"
    DefaultSchedulerInterval = "5m"
    DefaultKeepaliveTimeout  = "00:05:00"
)

// RouterOS paths
const (
    PathHotspotUser       = "/ip/hotspot/user"
    PathHotspotUserProfile = "/ip/hotspot/user/profile"
    PathHotspotActive      = "/ip/hotspot/active"
    PathSystemScheduler    = "/system/scheduler"
    PathSystemScript       = "/system/script"
)
```

## Error Handling

The library provides structured error handling:

```go
import "go-template/pkg/hotspot"

err := client.CreateUser(ctx, user)
if err != nil {
    if hotspot.IsNotFound(err) {
        // Handle not found
    } else if hotspot.IsInvalid(err) {
        // Handle invalid input
    } else if hErr, ok := err.(*hotspot.HotspotError); ok {
        // Access operation and context
        fmt.Printf("Operation %s failed: %v\n", hErr.Operation, hErr.Err)
    }
}
```

## Testing

The library uses interface-based design for easy testing:

```go
// Mock implementation
type MockRouterOSClient struct {
    // Your mock implementation
}

func (m *MockRouterOSClient) Run(args ...string) (*routeros.Reply, error) {
    // Return mock data
}

// Create client with mock
client := hotspot.NewClient(1, &MockRouterOSClient{})
```

## License

MIT
