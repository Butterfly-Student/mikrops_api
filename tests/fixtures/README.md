# Test Fixtures / Test Data Factory

This package provides a centralized test data factory for all test data needed across the project.

## Overview

The fixtures package offers comprehensive test data generation factories for:
- **Client** - API client test data
- **User** - User account test data
- **Mikrotik** - MikroTik router test data
- **PPPoE** - PPPoE session, secrets, profiles, and callback data
- **Queue** - Queue and queue statistics test data

## Usage

### Basic Usage

```go
import "go-template/tests/fixtures"

func TestSomething(t *testing.T) {
    factory := fixtures.NewTestDataFactory()
    
    // Client data
    client := factory.Client.ValidClient()
    
    // User data
    user := factory.User.ValidUser()
    
    // MikroTik data
    router := factory.Mikrotik.ValidMikrotikRouter()
    
    // PPPoE data
    secret := factory.Pppoe.ValidPppoeSecret()
    
    // Queue data
    queue := factory.Queue.ValidPppoeQueue()
}
```

### Or use individual factories

```go
import "go-template/tests/fixtures"

func TestUser(t *testing.T) {
    userFactory := fixtures.NewUserTestData()
    
    user := userFactory.ValidUser()
    admin := userFactory.AdminUser()
    inactive := userFactory.InactiveUser()
}
```

## Available Factories

### ClientTestData

Methods for generating client test data:
- `ValidClientInput()` - Valid client input for creation
- `ValidClient()` - Valid client model (ID: 1)
- `ValidClientFilter()` - Valid client filter
- `MultipleClients(count)` - Multiple client models
- `MultipleClientInputs(count)` - Multiple client inputs

Example:
```go
clients := factory.Client.MultipleClients(5)
```

### UserTestData

Methods for generating user test data:
- `ValidUserInput()` - Valid user input for creation
- `ValidUser()` - Valid user model (user role)
- `AdminUser()` - Admin user model
- `InactiveUser()` - Inactive user model
- `ValidUserFilter()` - Valid user filter
- `MultipleUsers(count)` - Multiple user models with mixed roles
- `MultipleUserInputs(count)` - Multiple user inputs
- `UserWithRole(role)` - User with specific role
- `UserWithStatus(status)` - User with specific status

Example:
```go
users := factory.User.MultipleUsers(10)
adminUser := factory.User.UserWithRole("admin")
inactiveUser := factory.User.UserWithStatus("inactive")
```

### MikrotikTestData

Methods for generating MikroTik router test data:
- `ValidMikrotikRouter()` - Valid router model
- `ValidMikrotikRouterInput()` - Valid router input
- `ActiveRouter()` - Active router
- `InactiveRouter()` - Inactive router
- `RouterWithSSL()` - Router with SSL enabled
- `MultipleRouters(count)` - Multiple routers (alternating active/inactive)
- `RouterWithAddress(address)` - Router with specific address
- `RouterWithName(name)` - Router with specific name

Example:
```go
routers := factory.Mikrotik.MultipleRouters(3)
activeRouters := factory.Mikrotik.ActiveRouter()
sslRouter := factory.Mikrotik.RouterWithSSL()
```

### PppoeTestData

Methods for generating PPPoE test data:
- `ValidPppoeSecret()` - Valid PPPoE secret
- `ValidPppoeProfile()` - Valid PPPoE profile
- `ValidPppoeActive()` - Valid active PPPoE session
- `ValidPppoeCallbackData()` - Valid callback data from RouterOS
- `MultiplePppoeSecrets(count)` - Multiple secrets
- `MultiplePppoeProfiles(count)` - Multiple profiles
- `MultiplePppoeActive(count)` - Multiple active sessions
- `PppoeSecretWithName(name)` - Secret with specific name
- `PppoeProfileWithName(name)` - Profile with specific name
- `PppoeActiveWithUser(username)` - Active session for user
- `PppoeCallbackDataWithUser(username)` - Callback data for user
- `WebSocketMessageSessionUp(username)` - Session up WebSocket message
- `WebSocketMessageSessionDown(username)` - Session down WebSocket message

Example:
```go
secrets := factory.Pppoe.MultiplePppoeSecrets(5)
sessionUp := factory.Pppoe.WebSocketMessageSessionUp("john-doe")
```

### QueueTestData

Methods for generating queue test data:
- `ValidPppoeQueue()` - Valid PPPoE queue
- `ValidQueueStats()` - Valid queue statistics
- `MultiplePppoeQueues(count)` - Multiple queues
- `MultipleQueueStats(count)` - Multiple queue statistics
- `PppoeQueueWithTarget(target)` - Queue for specific target
- `PppoeQueueWithName(name)` - Queue with specific name
- `PppoeQueueWithMaxLimit(maxLimit)` - Queue with specific max limit
- `PppoeQueueWithPriority(priority)` - Queue with specific priority
- `HighPriorityQueue()` - High priority queue (priority 1)
- `LowPriorityQueue()` - Low priority queue (priority 8)
- `QueueStatsWithName(name)` - Queue stats for specific queue

Example:
```go
queues := factory.Queue.MultiplePppoeQueues(10)
highPriority := factory.Queue.HighPriorityQueue()
customQueue := factory.Queue.PppoeQueueWithTarget("192.168.1.100/32")
```

## Best Practices

1. **Use the factory pattern** - Always use factories instead of hardcoding test data
2. **Reuse factories** - Create one factory instance per test suite or reuse across tests
3. **Customize when needed** - Use specific setter methods (`WithName()`, `WithRole()`, etc.) instead of modifying returned data
4. **Keep data realistic** - Use the provided factories as-is for most cases to test realistic scenarios
5. **Document custom data** - If creating custom data, add comments explaining what you're testing

## Test Data Conventions

- **Email format**: `{entity}{n}@example.com` (e.g., `user1@example.com`, `admin@example.com`)
- **Passwords**: Follow format `{Type}Password@123` (e.g., `TestPassword@123`, `AdminPassword@123`)
- **Timestamps**: Use `time.Now()` for most cases or specify in custom methods
- **IDs**: Start from 1 and increment sequentially
- **Statuses**: "active" or "inactive"
- **Roles**: "user" or "admin"

## Integration with Tests

### Unit Tests
```go
func TestUserService(t *testing.T) {
    factory := fixtures.NewTestDataFactory()
    user := factory.User.ValidUser()
    
    // Use user in test
    service := NewUserService(mockRepo)
    result := service.GetUser(user.ID)
}
```

### Integration Tests
```go
func TestUserIntegration(t *testing.T) {
    factory := fixtures.NewTestDataFactory()
    users := factory.User.MultipleUsers(3)
    
    // Insert test data
    for _, user := range users {
        db.Create(&user)
    }
    
    // Run integration test
    adapter := postgres.NewUserAdapter(db)
    found := adapter.FindByEmail(users[0].Email)
}
```

## Adding New Fixtures

To add test data for new entities:

1. Create a new file in `tests/fixtures/` (e.g., `newentity.go`)
2. Define a `NewEntityTestData` struct
3. Implement factory methods following the same pattern
4. Add the factory to `TestDataFactory` in `factory.go`

Example:
```go
// tests/fixtures/myentity.go
package fixtures

type MyEntityTestData struct{}

func NewMyEntityTestData() *MyEntityTestData {
    return &MyEntityTestData{}
}

func (m *MyEntityTestData) ValidEntity() model.MyEntity {
    // Implementation
}
```

```go
// Update factory.go
type TestDataFactory struct {
    MyEntity *MyEntityTestData
    // ... other factories
}

func NewTestDataFactory() *TestDataFactory {
    return &TestDataFactory{
        MyEntity: NewMyEntityTestData(),
        // ... other factories
    }
}
```
