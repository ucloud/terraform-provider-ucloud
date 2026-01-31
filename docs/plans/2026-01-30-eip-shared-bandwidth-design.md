# EIP Shared Bandwidth Package Support Design

**Date**: 2026-01-30
**Status**: Approved
**Author**: Design collaboration with user

## Overview

Add `share_bandwidth_package_id` support to the `ucloud_eip` resource, allowing EIPs to join or leave shared bandwidth packages. The implementation will support three operations:

- **Create with shared bandwidth**: Allocate EIP directly into a bandwidth package
- **Associate**: Add an existing EIP to a shared bandwidth package
- **Disassociate**: Remove an EIP from a shared bandwidth package

## Current State

- Documentation already mentions `share_bandwidth_package_id` parameter
- Schema doesn't include this field
- Create/Update/Read functions don't handle shared bandwidth
- UCloud SDK has full support via `AllocateEIP`, `AssociateEIPWithShareBandwidth`, and `DisassociateEIPWithShareBandwidth` APIs

## Design Decisions

### 1. Update Support
**Decision**: Allow updates (bind/unbind operations)
**Rationale**: Users need flexibility to move EIPs in and out of shared bandwidth packages without recreating resources.

### 2. Removal Behavior
**Decision**: Require explicit bandwidth and charge_mode values when removing shared bandwidth
**Rationale**: Forces users to think about the EIP configuration after leaving shared bandwidth, preventing accidental invalid states.

### 3. Validation Strategy
**Decision**: Strict validation with clear error messages
**Rationale**: Prevents API errors by catching invalid configurations at plan time.

### 4. State Storage
**Decision**: Store only the package ID, not additional attributes
**Rationale**: Keeps it simple and focused on user configuration needs.

## Implementation Details

### Schema Changes

Add to `resource_ucloud_eip.go` schema (after line 63):

```go
"share_bandwidth_package_id": {
    Type:     schema.TypeString,
    Optional: true,
},
```

Update `charge_mode` validation to include "share_bandwidth":

```go
"charge_mode": {
    Type:     schema.TypeString,
    Optional: true,
    Computed: true,
    ValidateFunc: validation.StringInSlice([]string{
        "traffic",
        "bandwidth",
        "share_bandwidth",
    }, false),
},
```

### Validation Function

Create new validation function:

```go
func validateSharedBandwidthConfig(d *schema.ResourceData) error {
    shareBandwidthId, hasShareBandwidth := d.GetOk("share_bandwidth_package_id")
    bandwidth := d.Get("bandwidth").(int)
    chargeMode := d.Get("charge_mode").(string)

    if hasShareBandwidth && shareBandwidthId.(string) != "" {
        // Has shared bandwidth - strict requirements
        if chargeMode != "share_bandwidth" {
            return fmt.Errorf("charge_mode must be 'share_bandwidth' when share_bandwidth_package_id is set, got '%s'", chargeMode)
        }
        if bandwidth != 0 {
            return fmt.Errorf("bandwidth must be 0 when share_bandwidth_package_id is set, got %d", bandwidth)
        }
    } else {
        // No shared bandwidth - regular requirements
        if chargeMode == "share_bandwidth" {
            return fmt.Errorf("charge_mode cannot be 'share_bandwidth' without share_bandwidth_package_id")
        }
        if bandwidth == 0 {
            return fmt.Errorf("bandwidth must be greater than 0 when not using shared bandwidth package")
        }
    }

    return nil
}
```

### Create Logic

In `resourceUCloudEIPCreate`, add validation before API call (around line 156):

```go
// Validate shared bandwidth configuration
if err := validateSharedBandwidthConfig(d); err != nil {
    return err
}
```

After bandwidth setup (around line 175), add:

```go
if v, ok := d.GetOk("share_bandwidth_package_id"); ok {
    req.ShareBandwidthId = ucloud.String(v.(string))
}
```

### Update Logic

In `resourceUCloudEIPUpdate`, add validation at start:

```go
// Validate shared bandwidth configuration
if err := validateSharedBandwidthConfig(d); err != nil {
    return err
}
```

Add new update block after charge_mode update (around line 269):

```go
if d.HasChange("share_bandwidth_package_id") && !d.IsNewResource() {
    oldVal, newVal := d.GetChange("share_bandwidth_package_id")
    oldId := oldVal.(string)
    newId := newVal.(string)

    // Disassociate from old shared bandwidth package
    if oldId != "" {
        reqDisassoc := conn.NewDisassociateEIPWithShareBandwidthRequest()
        reqDisassoc.ShareBandwidthId = ucloud.String(oldId)
        reqDisassoc.EIPIds = []string{d.Id()}
        reqDisassoc.Bandwidth = ucloud.Int(d.Get("bandwidth").(int))
        reqDisassoc.PayMode = ucloud.String(upperCamelCvt.unconvert(d.Get("charge_mode").(string)))

        _, err := conn.DisassociateEIPWithShareBandwidth(reqDisassoc)
        if err != nil {
            return fmt.Errorf("error on disassociating eip %q from shared bandwidth, %s", d.Id(), err)
        }

        // Wait for disassociation to complete
        stateConf := eipWaitForState(client, d.Id())
        _, err = stateConf.WaitForState()
        if err != nil {
            return fmt.Errorf("error on waiting for disassociation complete for eip %q, %s", d.Id(), err)
        }
    }

    // Associate with new shared bandwidth package
    if newId != "" {
        reqAssoc := conn.NewAssociateEIPWithShareBandwidthRequest()
        reqAssoc.ShareBandwidthId = ucloud.String(newId)
        reqAssoc.EIPIds = []string{d.Id()}

        _, err := conn.AssociateEIPWithShareBandwidth(reqAssoc)
        if err != nil {
            return fmt.Errorf("error on associating eip %q with shared bandwidth, %s", d.Id(), err)
        }

        // Wait for association to complete
        stateConf := eipWaitForState(client, d.Id())
        _, err = stateConf.WaitForState()
        if err != nil {
            return fmt.Errorf("error on waiting for association complete for eip %q, %s", d.Id(), err)
        }
    }

    d.SetPartial("share_bandwidth_package_id")
}
```

### Read Logic

In `resourceUCloudEIPRead`, after existing field reads (around line 340):

```go
// Set share_bandwidth_package_id from ShareBandwidthSet
if eip.ShareBandwidthSet.ShareBandwidthId != "" {
    d.Set("share_bandwidth_package_id", eip.ShareBandwidthSet.ShareBandwidthId)
} else {
    d.Set("share_bandwidth_package_id", "")
}
```

## Testing Strategy

### Test Cases

1. **TestAccUCloudEIP_sharedbandwidth_basic**
   - Create EIP with shared bandwidth package
   - Verify correct association

2. **TestAccUCloudEIP_sharedbandwidth_update**
   - Create EIP without shared bandwidth
   - Update to add shared bandwidth
   - Update to remove shared bandwidth
   - Verify state changes at each step

3. **TestAccUCloudEIP_sharedbandwidth_validation**
   - Test validation errors for invalid configurations

### Test Prerequisites

Tests require a shared bandwidth package to exist. Options:
- Create package as part of test setup
- Use environment variable for pre-existing package ID
- Document requirement for testers

## Documentation Updates

The documentation at `website/docs/r/eip.html.markdown` already includes `share_bandwidth_package_id` on line 34. We should:

1. Verify the description is accurate
2. Add example showing shared bandwidth usage:

```hcl
resource "ucloud_eip" "example_shared" {
  internet_type               = "bgp"
  bandwidth                   = 0
  charge_mode                 = "share_bandwidth"
  share_bandwidth_package_id  = "bwpack-xxx"
  name                        = "tf-example-eip-shared"
  tag                         = "tf-example"
}
```

## API References

### UCloud SDK APIs Used

- `AllocateEIP`: Create EIP (with optional ShareBandwidthId parameter)
- `AssociateEIPWithShareBandwidth`: Add EIP to shared bandwidth package
- `DisassociateEIPWithShareBandwidth`: Remove EIP from shared bandwidth package
- `DescribeEIP`: Returns EIP with ShareBandwidthSet containing package details

### Data Structures

- `AllocateEIPRequest.ShareBandwidthId`: Optional field for create-time association
- `AssociateEIPWithShareBandwidthRequest`: Requires EIPIds and ShareBandwidthId
- `DisassociateEIPWithShareBandwidthRequest`: Requires EIPIds, ShareBandwidthId, Bandwidth, and PayMode
- `UnetEIPSet.ShareBandwidthSet`: Contains ShareBandwidthId, ShareBandwidth, and ShareBandwidthName

## Validation Rules

| Condition | charge_mode | bandwidth | Valid? |
|-----------|-------------|-----------|--------|
| With share_bandwidth_package_id | "share_bandwidth" | 0 | ✅ Yes |
| With share_bandwidth_package_id | "bandwidth" or "traffic" | any | ❌ No |
| With share_bandwidth_package_id | "share_bandwidth" | >0 | ❌ No |
| Without share_bandwidth_package_id | "bandwidth" or "traffic" | >0 | ✅ Yes |
| Without share_bandwidth_package_id | "share_bandwidth" | any | ❌ No |
| Without share_bandwidth_package_id | any | 0 | ❌ No |

## Implementation Order

1. Add schema field and update charge_mode validation
2. Implement validation function
3. Update create logic with validation and ShareBandwidthId parameter
4. Update read logic to populate share_bandwidth_package_id
5. Implement update logic with associate/disassociate operations
6. Add tests
7. Update documentation with examples

## Edge Cases

1. **Import**: Works automatically via ImportStatePassthrough and read logic
2. **Empty package ID**: Treated as no shared bandwidth (disassociation)
3. **Changing package ID**: Disassociates from old, associates with new
4. **API failures**: Return errors with context about which operation failed
5. **State waiting**: Use existing eipWaitForState after each operation
