# EIP Shared Bandwidth Package Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `share_bandwidth_package_id` support to `ucloud_eip` resource to enable EIPs to join/leave shared bandwidth packages.

**Architecture:** Extend existing EIP resource with new schema field, validation logic, and update handlers for associate/disassociate operations using UCloud SDK APIs.

**Tech Stack:** Go, Terraform Plugin SDK, UCloud Go SDK

---

## Task 1: Add Schema Field and Update Validation

**Files:**
- Modify: `ucloud/resource_ucloud_eip.go:25-150`

**Step 1: Add share_bandwidth_package_id to schema**

Add after the `charge_mode` field (around line 63):

```go
"share_bandwidth_package_id": {
    Type:     schema.TypeString,
    Optional: true,
},
```

**Step 2: Update charge_mode validation to include share_bandwidth**

Find the `charge_mode` field definition (around line 55) and update:

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

**Step 3: Verify compilation**

Run: `go build ./ucloud`
Expected: SUCCESS (builds without errors)

**Step 4: Commit**

```bash
git add ucloud/resource_ucloud_eip.go
git commit -m "feat: add share_bandwidth_package_id schema field

Add share_bandwidth_package_id field to EIP resource schema and
update charge_mode validation to include share_bandwidth mode.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Implement Validation Function

**Files:**
- Modify: `ucloud/resource_ucloud_eip.go:416+` (add at end of file)

**Step 1: Add validation function**

Add this function at the end of the file (after `eipWaitForState`):

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

**Step 2: Verify compilation**

Run: `go build ./ucloud`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add ucloud/resource_ucloud_eip.go
git commit -m "feat: add shared bandwidth validation function

Add validateSharedBandwidthConfig to enforce configuration rules:
- When share_bandwidth_package_id is set: charge_mode must be
  'share_bandwidth' and bandwidth must be 0
- When not set: charge_mode cannot be 'share_bandwidth' and
  bandwidth must be > 0

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 3: Update Create Logic

**Files:**
- Modify: `ucloud/resource_ucloud_eip.go:153-220`

**Step 1: Add validation call at start of resourceUCloudEIPCreate**

Add after line 156 (after `req := conn.NewAllocateEIPRequest()`):

```go
// Validate shared bandwidth configuration
if err := validateSharedBandwidthConfig(d); err != nil {
    return err
}
```

**Step 2: Add ShareBandwidthId to request**

Add after the remark field handling (around line 198):

```go
if v, ok := d.GetOk("share_bandwidth_package_id"); ok {
    req.ShareBandwidthId = ucloud.String(v.(string))
}
```

**Step 3: Verify compilation**

Run: `go build ./ucloud`
Expected: SUCCESS

**Step 4: Run unit tests**

Run: `go test ./ucloud -short -run TestAccUCloudEIP`
Expected: PASS (tests still pass with changes)

**Step 5: Commit**

```bash
git add ucloud/resource_ucloud_eip.go
git commit -m "feat: add shared bandwidth support to EIP create

Add validation and ShareBandwidthId parameter to EIP creation,
allowing EIPs to be allocated directly into shared bandwidth packages.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 4: Update Read Logic

**Files:**
- Modify: `ucloud/resource_ucloud_eip.go:320-365`

**Step 1: Add share_bandwidth_package_id to read function**

Add after the resource field setting (around line 362, before the final return):

```go
// Set share_bandwidth_package_id from ShareBandwidthSet
if eip.ShareBandwidthSet.ShareBandwidthId != "" {
    d.Set("share_bandwidth_package_id", eip.ShareBandwidthSet.ShareBandwidthId)
} else {
    d.Set("share_bandwidth_package_id", "")
}
```

**Step 2: Verify compilation**

Run: `go build ./ucloud`
Expected: SUCCESS

**Step 3: Run unit tests**

Run: `go test ./ucloud -short`
Expected: PASS

**Step 4: Commit**

```bash
git add ucloud/resource_ucloud_eip.go
git commit -m "feat: read shared bandwidth package ID from API

Read ShareBandwidthSet.ShareBandwidthId from EIP response and
populate share_bandwidth_package_id in Terraform state.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 5: Implement Update Logic

**Files:**
- Modify: `ucloud/resource_ucloud_eip.go:222-318`

**Step 1: Add validation at start of resourceUCloudEIPUpdate**

Add after line 226 (after `d.Partial(true)`):

```go
// Validate shared bandwidth configuration
if err := validateSharedBandwidthConfig(d); err != nil {
    return err
}
```

**Step 2: Add shared bandwidth update handler**

Add after the charge_mode update block (around line 269):

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

        d.SetPartial("share_bandwidth_package_id")

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

        d.SetPartial("share_bandwidth_package_id")

        // Wait for association to complete
        stateConf := eipWaitForState(client, d.Id())
        _, err = stateConf.WaitForState()
        if err != nil {
            return fmt.Errorf("error on waiting for association complete for eip %q, %s", d.Id(), err)
        }
    }
}
```

**Step 3: Verify compilation**

Run: `go build ./ucloud`
Expected: SUCCESS

**Step 4: Run unit tests**

Run: `go test ./ucloud -short`
Expected: PASS

**Step 5: Commit**

```bash
git add ucloud/resource_ucloud_eip.go
git commit -m "feat: add shared bandwidth update support

Implement associate/disassociate operations in EIP update handler,
allowing EIPs to be moved in and out of shared bandwidth packages.
Includes state waiting for operation completion.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 6: Update Documentation Example

**Files:**
- Modify: `website/docs/r/eip.html.markdown:9-23`

**Step 1: Add shared bandwidth example**

Add after the existing basic example (around line 23):

```markdown
## Example Usage with Shared Bandwidth

```hcl
resource "ucloud_eip" "example_shared" {
  internet_type              = "bgp"
  bandwidth                  = 0
  charge_mode                = "share_bandwidth"
  share_bandwidth_package_id = "bwpack-xxxxx"
  name                       = "tf-example-eip-shared"
  tag                        = "tf-example"
}
```
```

**Step 2: Verify markdown formatting**

Run: `cat website/docs/r/eip.html.markdown | grep -A 10 "Shared Bandwidth"`
Expected: Shows properly formatted example

**Step 3: Commit**

```bash
git add website/docs/r/eip.html.markdown
git commit -m "docs: add shared bandwidth example to EIP resource

Add example configuration showing how to create an EIP with
shared bandwidth package.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 7: Add Basic Unit Test for Validation

**Files:**
- Create: `ucloud/resource_ucloud_eip_shared_bandwidth_test.go`

**Step 1: Create test file with validation tests**

```go
package ucloud

import (
    "testing"

    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestValidateSharedBandwidthConfig_WithPackageId(t *testing.T) {
    d := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, map[string]interface{}{
        "internet_type":               "bgp",
        "share_bandwidth_package_id":  "bwpack-test",
        "charge_mode":                 "share_bandwidth",
        "bandwidth":                   0,
    })

    err := validateSharedBandwidthConfig(d)
    if err != nil {
        t.Fatalf("expected no error with valid shared bandwidth config, got: %s", err)
    }
}

func TestValidateSharedBandwidthConfig_WithPackageId_InvalidChargeMode(t *testing.T) {
    d := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, map[string]interface{}{
        "internet_type":               "bgp",
        "share_bandwidth_package_id":  "bwpack-test",
        "charge_mode":                 "bandwidth",
        "bandwidth":                   0,
    })

    err := validateSharedBandwidthConfig(d)
    if err == nil {
        t.Fatal("expected error with invalid charge_mode, got nil")
    }
}

func TestValidateSharedBandwidthConfig_WithPackageId_InvalidBandwidth(t *testing.T) {
    d := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, map[string]interface{}{
        "internet_type":               "bgp",
        "share_bandwidth_package_id":  "bwpack-test",
        "charge_mode":                 "share_bandwidth",
        "bandwidth":                   10,
    })

    err := validateSharedBandwidthConfig(d)
    if err == nil {
        t.Fatal("expected error with non-zero bandwidth, got nil")
    }
}

func TestValidateSharedBandwidthConfig_WithoutPackageId(t *testing.T) {
    d := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, map[string]interface{}{
        "internet_type": "bgp",
        "charge_mode":   "bandwidth",
        "bandwidth":     2,
    })

    err := validateSharedBandwidthConfig(d)
    if err != nil {
        t.Fatalf("expected no error with valid regular config, got: %s", err)
    }
}

func TestValidateSharedBandwidthConfig_WithoutPackageId_InvalidChargeMode(t *testing.T) {
    d := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, map[string]interface{}{
        "internet_type": "bgp",
        "charge_mode":   "share_bandwidth",
        "bandwidth":     2,
    })

    err := validateSharedBandwidthConfig(d)
    if err == nil {
        t.Fatal("expected error with share_bandwidth mode without package_id, got nil")
    }
}

func TestValidateSharedBandwidthConfig_WithoutPackageId_ZeroBandwidth(t *testing.T) {
    d := schema.TestResourceDataRaw(t, resourceUCloudEIP().Schema, map[string]interface{}{
        "internet_type": "bgp",
        "charge_mode":   "bandwidth",
        "bandwidth":     0,
    })

    err := validateSharedBandwidthConfig(d)
    if err == nil {
        t.Fatal("expected error with zero bandwidth without package_id, got nil")
    }
}
```

**Step 2: Run the validation tests**

Run: `go test ./ucloud -run TestValidateSharedBandwidthConfig -v`
Expected: PASS (all 6 tests pass)

**Step 3: Commit**

```bash
git add ucloud/resource_ucloud_eip_shared_bandwidth_test.go
git commit -m "test: add validation unit tests for shared bandwidth

Add comprehensive unit tests for validateSharedBandwidthConfig
covering all validation rules and edge cases.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 8: Final Verification

**Files:**
- Test: All modified files

**Step 1: Run all unit tests**

Run: `go test ./ucloud -short`
Expected: PASS (all tests pass)

**Step 2: Verify build**

Run: `go build ./...`
Expected: SUCCESS (builds without errors or warnings)

**Step 3: Run go vet**

Run: `go vet ./ucloud`
Expected: No issues reported

**Step 4: Run gofmt**

Run: `gofmt -l ucloud/`
Expected: No output (all files properly formatted)

**Step 5: Create summary commit if needed**

If all checks pass, optionally create a summary commit:

```bash
git commit --allow-empty -m "feat: complete EIP shared bandwidth support

Summary of changes:
- Added share_bandwidth_package_id schema field
- Implemented validation for shared bandwidth configurations
- Added create/read/update support for shared bandwidth operations
- Added unit tests for validation logic
- Updated documentation with example

Closes implementation of shared bandwidth package support for EIP resource.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Notes for Implementation

### Testing Considerations

**Unit Tests**: Implemented in Task 7, covers validation logic comprehensively.

**Acceptance Tests**: Not included in this plan as they require:
- A real UCloud account with credentials
- A pre-existing shared bandwidth package
- TF_ACC=1 environment variable

Acceptance tests should be added later following the existing pattern in `resource_ucloud_eip_test.go` when testing infrastructure is available.

### Validation Edge Cases Covered

1. ✅ With package ID: must have charge_mode="share_bandwidth" and bandwidth=0
2. ✅ Without package ID: cannot have charge_mode="share_bandwidth" or bandwidth=0
3. ✅ Empty string package ID: treated as "without package ID"

### API Operation Flow

**Create with shared bandwidth:**
1. Validate config
2. Call AllocateEIP with ShareBandwidthId
3. Wait for state
4. Read to populate state

**Update to add shared bandwidth:**
1. Validate new config
2. Call AssociateEIPWithShareBandwidth
3. Wait for state
4. Read to populate state

**Update to remove shared bandwidth:**
1. Validate new config (must have bandwidth > 0 and valid charge_mode)
2. Call DisassociateEIPWithShareBandwidth with new bandwidth and charge_mode
3. Wait for state
4. Read to populate state

**Update to change package:**
1. Validate new config
2. Disassociate from old package
3. Associate with new package
4. Read to populate state

### Common Issues to Watch For

- **Bandwidth value**: Must be 0 when using shared bandwidth, must be >0 otherwise
- **Charge mode**: Must be "share_bandwidth" when using shared bandwidth, cannot be "share_bandwidth" otherwise
- **State waiting**: Always wait for operations to complete before proceeding
- **Error handling**: Include EIP ID in error messages for debugging
- **Partial state**: Use d.SetPartial() to track field updates properly
