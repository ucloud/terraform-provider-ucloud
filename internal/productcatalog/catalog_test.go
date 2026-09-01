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
		if definition.apiManager.Product == "" || definition.apiManager.Package == "" {
			t.Fatalf("catalog product %q has incomplete API Manager identity: %#v", definition.name, definition.apiManager)
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

func TestAPIManagerIdentityFor(t *testing.T) {
	want := map[string]APIManagerIdentity{
		"iam":      {Product: "IAM", Package: "iam"},
		"ipsecvpn": {Product: "IPSecVPN", Package: "ipsecvpn"},
		"label":    {Product: "Label", Package: "label"},
		"uaccount": {Product: "UAccount", Package: "uaccount"},
		"uads":     {Product: "UDDoS", Package: "uddos"},
		"udb":      {Product: "UDB", Package: "udb"},
		"udisk":    {Product: "UDisk", Package: "udisk"},
		"udpn":     {Product: "UDPN", Package: "udpn"},
		"ufs":      {Product: "UFS", Package: "ufs"},
		"uhost":    {Product: "UHost", Package: "uhost"},
		"uk8s":     {Product: "UK8S", Package: "uk8s"},
		"ulb":      {Product: "ULB", Package: "ulb"},
		"umem":     {Product: "UMem", Package: "umem"},
		"unet":     {Product: "UNet", Package: "unet"},
		"uphost":   {Product: "UPHost", Package: "uphost"},
		"us3":      {Product: "UFile", Package: "ufile"},
		"vpc":      {Product: "VPC", Package: "vpc"},
	}
	if got := len(Names()); got != len(want) {
		t.Fatalf("historical product count = %d, mapped count = %d", got, len(want))
	}
	for name, expected := range want {
		got, ok := APIManagerIdentityFor(name)
		if !ok {
			t.Fatalf("APIManagerIdentityFor(%q) did not find historical product", name)
		}
		if got != expected {
			t.Errorf("APIManagerIdentityFor(%q) = %#v, want %#v", name, got, expected)
		}
	}
	if got, ok := APIManagerIdentityFor("unknown"); ok || got != (APIManagerIdentity{}) {
		t.Fatalf("APIManagerIdentityFor(unknown) = (%#v, %t), want zero, false", got, ok)
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
