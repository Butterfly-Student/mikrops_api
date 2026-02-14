# Phase 5: Mikrotik Integration (Isolation System) - Progress

## Overview
This document tracks the implementation progress of Phase 5: Mikrotik Integration with Isolation System for GoTemplate project.

## Status: IN PROGRESS (~30% Complete)

### ✅ Completed Features

#### 1. Customer Domain Methods
- ✅ **IsolateCustomer** - Isolate customer for payment overdue
  - Update customer status to "isolated"
  - Get isolation profile (limited speed bandwidth profile)
  - Update PPP secret on Mikrotik with isolation profile
  - Generate redirect script to payment portal
  - Apply firewall rules for walled garden

- ✅ **ActivateCustomer** - Activate customer after payment
  - Update customer status to "active"
  - Get original bandwidth profile
  - Restore PPP secret on Mikrotik with original profile
  - Remove isolation firewall rules
  - Send notification when available

#### 2. Helper Methods
- ✅ **GenerateRedirectScript** - Generate Mikrotik router script
  - Redirect HTTP/HTTPS to payment portal
  - Allow payment portal access
  - Customer-specific configuration

- ✅ **GenerateIsolationFirewallRules** - Generate firewall rules
  - NAT redirect rules
  - Portal allow rules
  - 4 rules total (HTTP/HTTPS redirect + accept)

- ✅ **getPaymentPortalURL** - Get payment portal URL
  - Can be configured via system settings
  - Default: "portal.example.com"

### ⚠️ In Progress / Issues

#### Domain Dependencies
The customer domain requires additional dependencies that need to be resolved:

```go
type domain struct {
	dbPort           outbound_port.DatabasePort
	mikrotikPort     outbound_port.MikrotikPort
	bandwidthProfile  bandwidth_profile.BandwidthProfileDomain  // NEEDS TO BE PASSED FROM REGISTRY
	notificationPort  notification.NotificationDomain            // NEEDS TO BE PASSED FROM REGISTRY
}
```

**Issues to Resolve:**
1. ❌ **Cyclic Dependency**: Cannot directly include `bandwidth_profile.BandwidthProfileDomain` in customer domain
2. ❌ **Cyclic Dependency**: Cannot directly include `notification.NotificationDomain` in customer domain
3. ❌ **MikrotikPort Missing Methods**: Need `AddFirewallRule` and `RemoveFirewallRule` methods
4. ❌ **Model Missing**: Need `model.FirewallRule` struct
5. ❌ **Missing Imports**: Need to add `fmt` and `log` to customer domain

#### Field Name Errors
Several field name errors in the implementation:
- ❌ `router.Host` → Should be `router.Address` (MikrotikRouter field)
- ❌ `customer.IPAddress` → Not available, use `router.Address` instead

### ❌ Not Started Features

#### 3. Firewall Rule Management
- Status: Not started
- Need to add methods to MikrotikPort
- Need to create FirewallRule model

#### 4. Mikrotik Router Integration
- Status: Not started
- Need to implement AddFirewallRule
- Need to implement RemoveFirewallRule

#### 5. Testing
- Status: Not started
- Unit tests for IsolateCustomer
- Unit tests for ActivateCustomer
- Integration tests for firewall rules

## API Endpoints (Already Implemented in Customer Handler)

```
POST   /customers/:id/isolate  # Isolate customer
POST   /customers/:id/activate # Activate customer
```

## Isolation Workflow

### Customer Isolation Flow
```
1. Admin requests isolation via API
   ↓
2. Domain updates customer status to "isolated"
   ↓
3. Get isolation bandwidth profile (limited speed)
   ↓
4. Update PPP secret on Mikrotik:
   - Set profile to isolation profile
   - Allow login (not disabled)
   ↓
5. Generate redirect script:
   - NAT rules to redirect HTTP/HTTPS to portal
   - Allow payment portal domain
   ↓
6. Apply firewall rules to Mikrotik
   ↓
7. Send notification to customer (optional)
   ↓
8. Customer redirected to payment portal on next HTTP request
   ↓
9. Customer makes payment
   ↓
10. Customer activated (flow below)
```

### Customer Activation Flow
```
1. Payment received
   ↓
2. Admin requests activation via API (or auto-activate)
   ↓
3. Domain updates customer status to "active"
   ↓
4. Get original bandwidth profile
   ↓
5. Restore PPP secret on Mikrotik:
   - Set profile back to original
   - Allow login
   ↓
6. Remove isolation firewall rules
   ↓
7. Send notification to customer (optional)
   ↓
8. Internet fully restored
```

## Generated Mikrotik Script

### Redirect Script Template
```mikrotik
# Customer: CUSTOMER_CODE
:local portalURL "portal.example.com"
:local customerCode "CUSTOMER_CODE"
# Redirect HTTP to portal
/ip firewall nat add chain=dstnat protocol=tcp dst-port=80 src-address=CUSTOMER_IP action=redirect to-ports=8080
# Walled garden (allow payment portal)
/ip firewall nat add chain=dstnat protocol=tcp dst-port=80,443 dst-address-list=portal-allowed action=accept
```

### Firewall Rules Generated
1. **HTTP Redirect Rule**
   - Chain: dstnat
   - Protocol: tcp
   - Src Address: Customer IP
   - Dst Port: 80
   - Action: redirect
   - To Ports: 8080 (Payment Portal)
   - Comment: Redirect CUSTOMER_CODE to payment portal

2. **HTTPS Redirect Rule**
   - Chain: dstnat
   - Protocol: tcp
   - Src Address: Customer IP
   - Dst Port: 443
   - Action: redirect
   - To Ports: 8080 (Payment Portal)
   - Comment: Redirect CUSTOMER_CODE HTTPS to payment portal

3. **HTTP Allow Portal Rule**
   - Chain: dstnat
   - Protocol: tcp
   - Dst Port: 80
   - Action: accept
   - Dst Address: portal.example.com
   - Comment: Allow payment portal access

4. **HTTPS Allow Portal Rule**
   - Chain: dstnat
   - Protocol: tcp
   - Dst Port: 443
   - Action: accept
   - Dst Address: portal.example.com
   - Comment: Allow payment portal HTTPS access

## Required Changes

### 1. Customer Domain Constructor
```go
func NewCustomerDomain(
	dbPort outbound_port.DatabasePort,
	mikrotikPort outbound_port.MikrotikPort,
	getBandwidthProfile func(ctx context.Context) bandwidth_profile.BandwidthProfileDomain,
	getNotificationDomain func(ctx context.Context) notification.NotificationDomain,
) CustomerDomain {
	return &domain{
		dbPort:           dbPort,
		mikrotikPort:     mikrotikPort,
		bandwidthProfile:  getBandwidthProfile,
		notificationPort:  getNotificationDomain,
	}
}
```

### 2. Add FirewallRule Model
```go
// internal/model/firewall_rule.go

type FirewallRule struct {
	Chain      string `json:"chain"`
	Protocol   string `json:"protocol"`
	SrcAddress string `json:"src_address"`
	DstPort    int    `json:"dst_port"`
	DstAddress string `json:"dst_address"`
	ToPorts    string `json:"to_ports"`
	Action     string `json:"action"`
	Comment    string `json:"comment"`
}
```

### 3. Add Methods to MikrotikPort
```go
// internal/port/outbound/mikrotik.go

type MikrotikPort interface {
	// ... existing methods
	
	AddFirewallRule(router *model.MikrotikRouter, rule model.FirewallRule) error
	RemoveFirewallRule(router *model.MikrotikRouter, rule model.FirewallRule) error
}
```

## Files Modified

### Already Modified
- `internal/domain/customer/domain.go` - Added IsolateCustomer, ActivateCustomer, helper methods
- `internal/port/inbound/customer.go` - Interface updated (IsolateCustomer, ActivateCustomer)
- `internal/adapter/inbound/gin/customer_handler.go` - Handler methods already exist

### Need to Modify
- `internal/domain/registry.go` - Update NewCustomerDomain to pass dependencies
- `internal/port/outbound/mikrotik.go` - Add firewall rule methods
- `internal/adapter/outbound/mikrotik/mikrotik_adapter.go` - Implement firewall rule methods
- `internal/model/firewall_rule.go` - CREATE NEW FILE
- `internal/migration/postgres/firewall_rules.go` - CREATE NEW MIGRATION

## Progress Summary

**Isolation System**: 🔄 ~50% Complete
- ✅ Domain logic (IsolateCustomer, ActivateCustomer)
- ✅ Helper methods (GenerateRedirectScript, GenerateIsolationFirewallRules)
- ✅ Customer handler methods
- ✅ Customer HTTP port
- ✅ Customer routes
- ⚠️ Dependencies (need to resolve cyclic dependency)
- ❌ Firewall rule management (not started)
- ❌ FirewallRule model (not created)
- ❌ Testing (not started)

**Overall Phase 5**: 🔄 ~23% Complete (1.5/6.5 features:
  - Notification System: ✅ 100%
  - Isolation System: 🔄 ~50%
  - Receipt Generation: ❌ 0%
  - Payment History: ❌ 0%
  - Refund Processing: ❌ 0%
  - Reconciliation: ❌ 0%)

## Next Steps

### Immediate (Priority 1)
1. **Resolve Cyclic Dependencies** - Update NewCustomerDomain to accept function parameters
2. **Create FirewallRule Model** - Add model for firewall rules
3. **Update MikrotikPort** - Add AddFirewallRule and RemoveFirewallRule methods
4. **Implement Firewall Methods** - In mikrotik adapter
5. **Test Isolation Flow** - End-to-end testing

### Short-term (Priority 2)
6. **Receipt Generation** - PDF receipt utility
7. **Payment History** - Customer payment history API

### Long-term (Priority 3)
8. **Refund Processing** - Refund workflow
9. **Reconciliation** - Reconciliation reports
10. **Multi-currency** - Currency support
11. **Subscription Management** - Auto-renewal, pause, cancel

## Known Issues

### LSP Errors
Multiple LSP errors in `internal/domain/customer/domain.go`:
- Missing imports (fmt, log)
- Field name errors (router.Host → router.Address)
- Missing dependencies (bandwidthProfile, notificationPort)
- MikrotikPort missing methods

### Test Files
Need to update test files that use `domain.NewDomain` after adding parameters.

---

**Last Updated**: 2025-02-14
**Next Task**: Resolve cyclic dependencies in customer domain
