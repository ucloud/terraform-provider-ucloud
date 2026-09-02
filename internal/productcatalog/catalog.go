// Package productcatalog owns the provider's static product composition.
package productcatalog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/terraform-providers/terraform-provider-ucloud/products/iam"
	"github.com/terraform-providers/terraform-provider-ucloud/products/ipsecvpn"
	"github.com/terraform-providers/terraform-provider-ucloud/products/label"
	"github.com/terraform-providers/terraform-provider-ucloud/products/uaccount"
	"github.com/terraform-providers/terraform-provider-ucloud/products/uads"
	"github.com/terraform-providers/terraform-provider-ucloud/products/udb"
	"github.com/terraform-providers/terraform-provider-ucloud/products/udisk"
	"github.com/terraform-providers/terraform-provider-ucloud/products/udpn"
	"github.com/terraform-providers/terraform-provider-ucloud/products/ufs"
	"github.com/terraform-providers/terraform-provider-ucloud/products/uhost"
	"github.com/terraform-providers/terraform-provider-ucloud/products/uk8s"
	"github.com/terraform-providers/terraform-provider-ucloud/products/ulb"
	"github.com/terraform-providers/terraform-provider-ucloud/products/umem"
	"github.com/terraform-providers/terraform-provider-ucloud/products/unet"
	"github.com/terraform-providers/terraform-provider-ucloud/products/uphost"
	"github.com/terraform-providers/terraform-provider-ucloud/products/us3"
	"github.com/terraform-providers/terraform-provider-ucloud/products/vpc"
)

// TestBaseline records the minimum compatibility coverage for one product.
type TestBaseline struct {
	AcceptanceTests int
	ImportTests     int
}

// ProductMasterDataIdentity records the canonical identity from product master data.
type ProductMasterDataIdentity struct {
	EnSampleName string
	Key          string
}

// OwnershipMetadata is the trusted product identity and Terraform surface used
// to derive product-owned repository paths.
type OwnershipMetadata struct {
	Name                string
	MasterData          ProductMasterDataIdentity
	TerraformNamespaces []string
	ResourceTypes       []string
	DataSourceTypes     []string
}

type definition struct {
	name                string
	masterData          ProductMasterDataIdentity
	newAdapter          func() product.V1
	terraformNamespaces []string
	testBaseline        TestBaseline
}

var definitions = []definition{
	{name: "iam", masterData: ProductMasterDataIdentity{EnSampleName: "IAM", Key: "iam"}, newAdapter: iam.New, testBaseline: TestBaseline{AcceptanceTests: 18}},
	{name: "ipsecvpn", masterData: ProductMasterDataIdentity{EnSampleName: "IPSecVPN", Key: "ipsecvpn"}, newAdapter: ipsecvpn.New, terraformNamespaces: []string{"vpn"}, testBaseline: TestBaseline{AcceptanceTests: 9, ImportTests: 3}},
	{name: "label", masterData: ProductMasterDataIdentity{EnSampleName: "Label", Key: "label"}, newAdapter: label.New, terraformNamespaces: []string{"label", "labels"}, testBaseline: TestBaseline{AcceptanceTests: 4}},
	{name: "uaccount", masterData: ProductMasterDataIdentity{EnSampleName: "UAccount", Key: "uaccount"}, newAdapter: uaccount.New, terraformNamespaces: []string{"projects", "zones"}, testBaseline: TestBaseline{AcceptanceTests: 2}},
	{name: "uads", masterData: ProductMasterDataIdentity{EnSampleName: "UDDoS", Key: "uddos"}, newAdapter: uads.New, terraformNamespaces: []string{"anti_ddos"}, testBaseline: TestBaseline{AcceptanceTests: 2, ImportTests: 1}},
	{name: "udb", masterData: ProductMasterDataIdentity{EnSampleName: "UDB", Key: "udb"}, newAdapter: udb.New, terraformNamespaces: []string{"db"}, testBaseline: TestBaseline{AcceptanceTests: 11, ImportTests: 1}},
	{name: "udisk", masterData: ProductMasterDataIdentity{EnSampleName: "UDisk", Key: "udisk"}, newAdapter: udisk.New, terraformNamespaces: []string{"disk", "disks"}, testBaseline: TestBaseline{AcceptanceTests: 16, ImportTests: 2}},
	{name: "udpn", masterData: ProductMasterDataIdentity{EnSampleName: "UDPN", Key: "udpn"}, newAdapter: udpn.New, testBaseline: TestBaseline{AcceptanceTests: 2, ImportTests: 1}},
	{name: "ufs", masterData: ProductMasterDataIdentity{EnSampleName: "UFS", Key: "ufs"}, newAdapter: ufs.New, testBaseline: TestBaseline{AcceptanceTests: 4}},
	{name: "uhost", masterData: ProductMasterDataIdentity{EnSampleName: "UHost", Key: "uhost"}, newAdapter: uhost.New, terraformNamespaces: []string{"instance", "instances", "images", "isolation_group"}, testBaseline: TestBaseline{AcceptanceTests: 20, ImportTests: 2}},
	{name: "uk8s", masterData: ProductMasterDataIdentity{EnSampleName: "UK8S", Key: "uk8s"}, newAdapter: uk8s.New, testBaseline: TestBaseline{AcceptanceTests: 2}},
	{name: "ulb", masterData: ProductMasterDataIdentity{EnSampleName: "ULB", Key: "ulb"}, newAdapter: ulb.New, terraformNamespaces: []string{"lb", "lbs"}, testBaseline: TestBaseline{AcceptanceTests: 21, ImportTests: 4}},
	{name: "umem", masterData: ProductMasterDataIdentity{EnSampleName: "UMem", Key: "umem"}, newAdapter: umem.New, terraformNamespaces: []string{"redis", "memcache"}, testBaseline: TestBaseline{AcceptanceTests: 3}},
	{name: "unet", masterData: ProductMasterDataIdentity{EnSampleName: "UNet", Key: "unet"}, newAdapter: unet.New, terraformNamespaces: []string{"eip", "eips", "security_group", "security_groups"}, testBaseline: TestBaseline{AcceptanceTests: 11, ImportTests: 2}},
	{name: "uphost", masterData: ProductMasterDataIdentity{EnSampleName: "UPHost", Key: "uphost"}, newAdapter: uphost.New, terraformNamespaces: []string{"baremetal"}, testBaseline: TestBaseline{AcceptanceTests: 3}},
	{name: "us3", masterData: ProductMasterDataIdentity{EnSampleName: "UFile", Key: "ufile"}, newAdapter: us3.New, testBaseline: TestBaseline{AcceptanceTests: 2}},
	{name: "vpc", masterData: ProductMasterDataIdentity{EnSampleName: "VPC", Key: "vpc"}, newAdapter: vpc.New, terraformNamespaces: []string{"vpc", "vpcs", "subnet", "subnets", "vip", "nat_gateway", "nat_gateways", "sec_group", "sec_groups"}, testBaseline: TestBaseline{AcceptanceTests: 16, ImportTests: 3}},
}

// Bindings returns fresh provider bindings for all products.
func Bindings() []product.Binding {
	bindings := make([]product.Binding, 0, len(definitions))
	for _, definition := range definitions {
		var options []product.BindOption
		if len(definition.terraformNamespaces) > 0 {
			options = append(options, product.WithTerraformNamespaces(definition.terraformNamespaces...))
		}
		bindings = append(bindings, product.Bind(definition.name, definition.newAdapter(), options...))
	}
	return bindings
}

// Names returns every registered product name in stable alphabetical order.
func Names() []string {
	names := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		names = append(names, definition.name)
	}
	return names
}

// ProductMasterDataIdentityFor resolves a stable Provider product name to product master data.
func ProductMasterDataIdentityFor(name string) (ProductMasterDataIdentity, bool) {
	for _, definition := range definitions {
		if definition.name == name {
			return definition.masterData, true
		}
	}
	return ProductMasterDataIdentity{}, false
}

// OwnershipMetadataFor returns the catalog and Registration data for one
// product. Product names must use the exact Provider catalog name.
func OwnershipMetadataFor(name string) (OwnershipMetadata, error) {
	for _, definition := range definitions {
		if definition.name != name {
			continue
		}
		registration := definition.newAdapter().Registration()
		if registration.Name != definition.name {
			return OwnershipMetadata{}, fmt.Errorf(
				"catalog product %q constructor registers %q",
				definition.name,
				registration.Name,
			)
		}
		namespaces := append([]string(nil), definition.terraformNamespaces...)
		if len(namespaces) == 0 {
			namespaces = []string{strings.ReplaceAll(definition.name, "-", "_")}
		}
		metadata := OwnershipMetadata{
			Name:                definition.name,
			MasterData:          definition.masterData,
			TerraformNamespaces: namespaces,
			ResourceTypes:       mapKeys(registration.Resources),
			DataSourceTypes:     mapKeys(registration.DataSources),
		}
		return metadata, nil
	}
	return OwnershipMetadata{}, fmt.Errorf(
		"unknown Provider product %q; choose one of: %s",
		name,
		strings.Join(Names(), ", "),
	)
}

// AllOwnershipMetadata returns trusted metadata for every catalog product.
func AllOwnershipMetadata() ([]OwnershipMetadata, error) {
	metadata := make([]OwnershipMetadata, 0, len(definitions))
	for _, definition := range definitions {
		entry, err := OwnershipMetadataFor(definition.name)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, entry)
	}
	return metadata, nil
}

// Baseline returns the compatibility test floor for a registered product.
func Baseline(name string) (TestBaseline, bool) {
	for _, definition := range definitions {
		if definition.name == name {
			return definition.testBaseline, true
		}
	}
	return TestBaseline{}, false
}

func mapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
