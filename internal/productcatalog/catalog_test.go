package productcatalog

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
)

func TestDefinitions(t *testing.T) {
	seen := make(map[string]struct{}, len(definitions))
	for index, definition := range definitions {
		if index > 0 && definitions[index-1].name >= definition.name {
			t.Fatalf("catalog product names must be unique and sorted: %q before %q", definitions[index-1].name, definition.name)
		}
		if _, exists := seen[definition.name]; exists {
			t.Fatalf("catalog contains duplicate product %q", definition.name)
		}
		seen[definition.name] = struct{}{}
		if definition.newAdapter == nil {
			t.Fatalf("catalog product %q has no adapter constructor", definition.name)
		}
		if definition.masterData.EnSampleName == "" || definition.masterData.Key == "" {
			t.Fatalf("catalog product %q has incomplete product master data identity: %#v", definition.name, definition.masterData)
		}
		adapter := definition.newAdapter()
		if adapter == nil {
			t.Fatalf("catalog product %q constructor returned nil", definition.name)
		}
		if got := adapter.Registration().Name; got != definition.name {
			t.Fatalf("catalog product %q constructor registers %q", definition.name, got)
		}
		if definition.testBaseline.AcceptanceTests <= 0 {
			t.Fatalf("catalog product %q must have a positive acceptance-test baseline", definition.name)
		}
		if definition.testBaseline.ImportTests < 0 || definition.testBaseline.ImportTests > definition.testBaseline.AcceptanceTests {
			t.Fatalf("catalog product %q has invalid import-test baseline %d", definition.name, definition.testBaseline.ImportTests)
		}
	}
}

func TestBindingsRegisterEveryProduct(t *testing.T) {
	bindings := Bindings()
	if got, want := len(bindings), len(definitions); got != want {
		t.Fatalf("binding count = %d, want %d", got, want)
	}
	provider := &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{},
		ResourcesMap:   map[string]*schema.Resource{},
	}
	if err := product.Register(provider, bindings...); err != nil {
		t.Fatalf("register catalog bindings: %v", err)
	}
}

func TestNamesReturnsCopy(t *testing.T) {
	want := []string{
		"iam", "ipsecvpn", "label", "uaccount", "uads", "udb", "udisk", "udpn", "ufs",
		"uhost", "uk8s", "ulb", "umem", "unet", "uphost", "us3", "vpc",
	}
	if got := Names(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Names() = %v, want historical names %v", got, want)
	}
	names := Names()
	names[0] = "changed"
	if got := Names()[0]; got == "changed" {
		t.Fatal("Names returned mutable catalog storage")
	}
}

func TestProductMasterDataIdentityFor(t *testing.T) {
	want := map[string]ProductMasterDataIdentity{
		"iam":      {EnSampleName: "IAM", Key: "iam"},
		"ipsecvpn": {EnSampleName: "IPSecVPN", Key: "ipsecvpn"},
		"label":    {EnSampleName: "Label", Key: "label"},
		"uaccount": {EnSampleName: "UAccount", Key: "uaccount"},
		"uads":     {EnSampleName: "UDDoS", Key: "uddos"},
		"udb":      {EnSampleName: "UDB", Key: "udb"},
		"udisk":    {EnSampleName: "UDisk", Key: "udisk"},
		"udpn":     {EnSampleName: "UDPN", Key: "udpn"},
		"ufs":      {EnSampleName: "UFS", Key: "ufs"},
		"uhost":    {EnSampleName: "UHost", Key: "uhost"},
		"uk8s":     {EnSampleName: "UK8S", Key: "uk8s"},
		"ulb":      {EnSampleName: "ULB", Key: "ulb"},
		"umem":     {EnSampleName: "UMem", Key: "umem"},
		"unet":     {EnSampleName: "UNet", Key: "unet"},
		"uphost":   {EnSampleName: "UPHost", Key: "uphost"},
		"us3":      {EnSampleName: "UFile", Key: "ufile"},
		"vpc":      {EnSampleName: "VPC", Key: "vpc"},
	}
	if got := len(Names()); got != len(want) {
		t.Fatalf("historical product count = %d, mapped count = %d", got, len(want))
	}
	for name, expected := range want {
		got, ok := ProductMasterDataIdentityFor(name)
		if !ok {
			t.Fatalf("ProductMasterDataIdentityFor(%q) did not find historical product", name)
		}
		if got != expected {
			t.Errorf("ProductMasterDataIdentityFor(%q) = %#v, want %#v", name, got, expected)
		}
	}
	if got, ok := ProductMasterDataIdentityFor("unknown"); ok || got != (ProductMasterDataIdentity{}) {
		t.Fatalf("ProductMasterDataIdentityFor(unknown) = (%#v, %t), want zero, false", got, ok)
	}
}

func TestBaselineLookup(t *testing.T) {
	for _, name := range Names() {
		if _, ok := Baseline(name); !ok {
			t.Fatalf("catalog product %q has no test baseline", name)
		}
	}
	if _, ok := Baseline("unknown"); ok {
		t.Fatal("unknown product has a test baseline")
	}
}
