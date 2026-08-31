package uphost

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestRegistrationKeepsLegacyTerraformSurface(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("baremetal"))); err != nil {
		t.Fatalf("register UPHost product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate UPHost provider: %v", err)
	}

	if provider.ResourcesMap["ucloud_baremetal_instance"] == nil {
		t.Fatal("ucloud_baremetal_instance is not registered")
	}
	if provider.DataSourcesMap["ucloud_baremetal_images"] == nil {
		t.Fatal("ucloud_baremetal_images is not registered")
	}
}

func TestBareMetalInstanceSchemaCompatibility(t *testing.T) {
	instance := resourceUCloudBareMetalInstance()
	if instance == nil {
		t.Fatal("ucloud_baremetal_instance is nil")
	}
	if instance.Create == nil || instance.Read == nil || instance.Update == nil || instance.Delete == nil {
		t.Fatal("ucloud_baremetal_instance CRUD callbacks are incomplete")
	}
	if instance.CustomizeDiff == nil {
		t.Fatal("ucloud_baremetal_instance CustomizeDiff is missing")
	}
	if instance.Importer == nil || instance.Importer.State == nil {
		t.Fatal("ucloud_baremetal_instance importer is missing")
	}
	if instance.Timeouts != nil {
		t.Fatal("ucloud_baremetal_instance unexpectedly added resource timeouts")
	}

	wantFields := map[string]struct {
		typeValue schema.ValueType
		required  bool
		optional  bool
		computed  bool
		forceNew  bool
		sensitive bool
	}{
		"availability_zone":           {typeValue: schema.TypeString, required: true, forceNew: true},
		"instance_type":               {typeValue: schema.TypeString, required: true, forceNew: true},
		"image_id":                    {typeValue: schema.TypeString, required: true},
		"allow_stopping_for_update":   {typeValue: schema.TypeBool, optional: true},
		"allow_stopping_for_resizing": {typeValue: schema.TypeBool, optional: true},
		"delete_disks_with_instance":  {typeValue: schema.TypeBool, optional: true, forceNew: true},
		"delete_eips_with_instance":   {typeValue: schema.TypeBool, optional: true, forceNew: true},
		"root_password":               {typeValue: schema.TypeString, optional: true, sensitive: true},
		"boot_disk_id":                {typeValue: schema.TypeString, computed: true},
		"boot_disk_size":              {typeValue: schema.TypeInt, optional: true},
		"boot_disk_type":              {typeValue: schema.TypeString, optional: true, forceNew: true},
		"charge_type":                 {typeValue: schema.TypeString, optional: true, forceNew: true},
		"duration":                    {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"name":                        {typeValue: schema.TypeString, optional: true},
		"remark":                      {typeValue: schema.TypeString, optional: true},
		"tag":                         {typeValue: schema.TypeString, optional: true},
		"security_group":              {typeValue: schema.TypeString, optional: true, computed: true},
		"vpc_id":                      {typeValue: schema.TypeString, required: true, forceNew: true},
		"subnet_id":                   {typeValue: schema.TypeString, required: true, forceNew: true},
		"private_ip":                  {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"data_disks":                  {typeValue: schema.TypeList, optional: true, forceNew: true},
		"network_interface":           {typeValue: schema.TypeList, optional: true, forceNew: true},
		"raid_type":                   {typeValue: schema.TypeString, optional: true, forceNew: true},
		"rdma_cluster_id":             {typeValue: schema.TypeString, computed: true},
	}
	assertSchemaFields(t, instance, wantFields)

	if instance.Schema["charge_type"].Default != "day" {
		t.Fatalf("charge_type default = %#v, want day", instance.Schema["charge_type"].Default)
	}
	if instance.Schema["duration"].Default != 1 {
		t.Fatalf("duration default = %#v, want 1", instance.Schema["duration"].Default)
	}
	if instance.Schema["tag"].Default != defaultTag {
		t.Fatalf("tag default = %#v, want %q", instance.Schema["tag"].Default, defaultTag)
	}
	if instance.Schema["data_disks"].MinItems != 0 || instance.Schema["data_disks"].MaxItems != 1 {
		t.Fatalf("data_disks cardinality = %d..%d, want 0..1", instance.Schema["data_disks"].MinItems, instance.Schema["data_disks"].MaxItems)
	}
	if instance.Schema["network_interface"].MinItems != 0 || instance.Schema["network_interface"].MaxItems != 1 {
		t.Fatalf("network_interface cardinality = %d..%d, want 0..1", instance.Schema["network_interface"].MinItems, instance.Schema["network_interface"].MaxItems)
	}

	dataDisks, ok := instance.Schema["data_disks"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("data_disks element type = %T, want *schema.Resource", instance.Schema["data_disks"].Elem)
	}
	assertSchemaFields(t, dataDisks, map[string]struct {
		typeValue schema.ValueType
		required  bool
		optional  bool
		computed  bool
		forceNew  bool
		sensitive bool
	}{
		"id":          {typeValue: schema.TypeString, computed: true},
		"device_name": {typeValue: schema.TypeString, computed: true},
		"size":        {typeValue: schema.TypeInt, required: true},
		"type":        {typeValue: schema.TypeString, required: true, forceNew: true},
	})

	interfaces, ok := instance.Schema["network_interface"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("network_interface element type = %T, want *schema.Resource", instance.Schema["network_interface"].Elem)
	}
	assertSchemaFields(t, interfaces, map[string]struct {
		typeValue schema.ValueType
		required  bool
		optional  bool
		computed  bool
		forceNew  bool
		sensitive bool
	}{
		"eip_bandwidth":     {typeValue: schema.TypeInt, required: true, forceNew: true},
		"eip_internet_type": {typeValue: schema.TypeString, required: true, forceNew: true},
		"eip_charge_mode":   {typeValue: schema.TypeString, required: true, forceNew: true},
	})
}

func TestBareMetalImagesDataSourceSchemaCompatibility(t *testing.T) {
	images := dataSourceUCloudBareMetalImages()
	if images == nil || images.Read == nil {
		t.Fatal("ucloud_baremetal_images Read callback is missing")
	}
	if len(images.Schema) != 10 {
		t.Fatalf("ucloud_baremetal_images schema has %d fields, want 10", len(images.Schema))
	}
	for _, name := range []string{
		"availability_zone", "name_regex", "image_type", "os_type", "image_id", "ids", "output_file", "instance_type", "images", "total_count",
	} {
		if images.Schema[name] == nil {
			t.Errorf("ucloud_baremetal_images field %q is missing", name)
		}
	}
	if field := images.Schema["ids"]; field.Type != schema.TypeSet || !field.Optional || !field.Computed || field.Set == nil {
		t.Fatal("ids schema no longer preserves optional computed TypeSet behavior")
	}
	if field := images.Schema["images"]; field.Type != schema.TypeList || !field.Computed {
		t.Fatal("images schema is not a computed TypeList")
	}
	nested, ok := images.Schema["images"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("images element type = %T, want *schema.Resource", images.Schema["images"].Elem)
	}
	for _, name := range []string{"id", "name", "type", "size", "availability_zone", "os_type", "os_name", "description", "status"} {
		if nested.Schema[name] == nil || !nested.Schema[name].Computed {
			t.Errorf("images nested field %q is not computed", name)
		}
	}
}

func TestClientComposesLegacyClientsAndTimeouts(t *testing.T) {
	config := &ucloud.Config{Region: "cn-bj2", ProjectId: "project-test", Timeout: 7 * time.Second}
	credential := &auth.Credential{PublicKey: "public-test", PrivateKey: "private-test"}
	handler := ucloud.HttpRequestHandler(func(_ *ucloud.Client, req *http.HttpRequest) (*http.HttpRequest, error) {
		return req, nil
	})

	rawClient := newClient(config, credential, []ucloud.HttpRequestHandler{handler})
	client, ok := rawClient.(*productClient)
	if !ok {
		t.Fatalf("newClient() returned %T, want *productClient", rawClient)
	}
	if client.uphostconn == nil || client.unetconn == nil || client.genericClient == nil {
		t.Fatal("UPHost product did not compose all legacy SDK clients")
	}
	if client.region != config.Region || client.projectId != config.ProjectId {
		t.Fatalf("product client identity = (%q, %q), want (%q, %q)", client.region, client.projectId, config.Region, config.ProjectId)
	}
	if client.unetconn.GetConfig() != config {
		t.Fatal("UNet client no longer uses the provider config")
	}
	if got := client.unetconn.GetConfig().Timeout; got != config.Timeout {
		t.Fatalf("UNet timeout = %s, want caller timeout %s", got, config.Timeout)
	}
	if got := client.uphostconn.GetConfig().Timeout; got != 60*time.Second {
		t.Fatalf("UPHost timeout = %s, want 1m0s", got)
	}
	if got := client.genericClient.GetConfig().Timeout; got != 60*time.Second {
		t.Fatalf("generic client timeout = %s, want 1m0s", got)
	}
	for name, sdkConfig := range map[string]*ucloud.Config{
		"UPHost":  client.uphostconn.GetConfig(),
		"UNet":    client.unetconn.GetConfig(),
		"generic": client.genericClient.GetConfig(),
	} {
		if sdkConfig.Region != config.Region || sdkConfig.ProjectId != config.ProjectId {
			t.Errorf("%s config = region %q project %q, want region %q project %q", name, sdkConfig.Region, sdkConfig.ProjectId, config.Region, config.ProjectId)
		}
	}
	if config.Timeout != 7*time.Second {
		t.Fatalf("newClient changed caller timeout to %s", config.Timeout)
	}
	for name, sdkClient := range map[string]interface{}{
		"UPHost":  client.uphostconn,
		"UNet":    client.unetconn,
		"generic": client.genericClient,
	} {
		if got := httpRequestHandlerCount(sdkClient); got != 1 {
			t.Errorf("%s custom HTTP handler count = %d, want 1", name, got)
		}
	}
	if client.uphostconn.GetCredential() != credential || client.unetconn.GetCredential() != credential || client.genericClient.GetCredential() != credential {
		t.Fatal("client composition did not preserve credential pointer")
	}
}

func TestClientFromMetaUsesProductRuntime(t *testing.T) {
	runtime := &testRuntime{config: &ucloud.Config{Region: "cn-test", ProjectId: "project-test"}}
	client, err := clientFromMeta(runtime)
	if err != nil {
		t.Fatalf("clientFromMeta() error = %v", err)
	}
	if runtime.name != Name || runtime.calls != 1 {
		t.Fatalf("runtime calls = name %q calls %d, want name %q calls 1", runtime.name, runtime.calls, Name)
	}
	if client == nil || client.uphostconn == nil || client.unetconn == nil || client.genericClient == nil {
		t.Fatal("clientFromMeta did not return a complete product client")
	}
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("clientFromMeta accepted a non-runtime value")
	}
}

func TestCallbacksRejectInvalidProviderRuntime(t *testing.T) {
	invalidMeta := struct{}{}
	callbacks := map[string]func() error{
		"create": func() error {
			return resourceUCloudBareMetalInstanceCreate(nil, invalidMeta)
		},
		"read": func() error {
			return resourceUCloudBareMetalInstanceRead(nil, invalidMeta)
		},
		"update": func() error {
			return resourceUCloudBareMetalInstanceUpdate(nil, invalidMeta)
		},
		"delete": func() error {
			return resourceUCloudBareMetalInstanceDelete(nil, invalidMeta)
		},
		"customize diff": func() error {
			return bareMetalInstanceCustomizeDiff(nil, invalidMeta)
		},
		"images data source": func() error {
			return dataSourceUCloudBareMetalImagesRead(nil, invalidMeta)
		},
	}

	for name, callback := range callbacks {
		t.Run(name, func(t *testing.T) {
			err := callback()
			if err == nil || !strings.Contains(err.Error(), "invalid provider runtime") {
				t.Fatalf("callback error = %v, want invalid provider runtime", err)
			}
		})
	}
}

func TestBareMetalCompatibilityHelpers(t *testing.T) {
	instance := resourceUCloudBareMetalInstance()
	for input, want := range map[string]string{
		"Year":               "year",
		"Month":              "month",
		"DescribePHostImage": "describe_p_host_image",
	} {
		if got := upperCamelCvt.convert(input); got != want {
			t.Errorf("upperCamelCvt.convert(%q) = %q, want %q", input, got, want)
		}
	}
	for input, want := range map[string]string{
		"year":                  "Year",
		"month":                 "Month",
		"describe_p_host_image": "DescribePHostImage",
	} {
		if got := upperCamelCvt.unconvert(input); got != want {
			t.Errorf("upperCamelCvt.unconvert(%q) = %q, want %q", input, got, want)
		}
	}
	if got := raidTypeCvt.unconvert("no_raid"); got != "NoRaid" {
		t.Fatalf("raidTypeCvt.unconvert(no_raid) = %q, want NoRaid", got)
	}
	if got := upperCvt.convert("International"); got != "international" {
		t.Fatalf("upperCvt.convert(International) = %q, want international", got)
	}
	if got := stateFuncTag(""); got != defaultTag {
		t.Fatalf("empty tag state = %q, want %q", got, defaultTag)
	}
	for value, wantErr := range map[string]bool{
		"test1234": false,
		"bad-pass": true,
		"short":    true,
	} {
		_, errors := validateUcloudInstanceRootPassword(value, "root_password")
		if got := len(errors) > 0; got != wantErr {
			t.Errorf("root password validation for %q = %v, wantErr = %t", value, errors, wantErr)
		}
	}
	if err := newNotFoundError(getNotFoundMessage("instance", "instance-test")); !isNotFoundError(err) {
		t.Fatal("new not-found error is not recognized")
	} else if got := err.Error(); !strings.Contains(got, "the specified instance instance-test is not found") {
		t.Fatalf("not-found error = %q, want legacy message", got)
	}

	legacy := &terraform.InstanceState{
		ID: "baremetal-legacy",
		Attributes: map[string]string{
			"name": "legacy-baremetal",
			"tag":  defaultTag,
		},
	}
	state := resourceUCloudBareMetalInstance().Data(legacy).State()
	if state == nil || state.ID != legacy.ID {
		t.Fatal("legacy baremetal state was not retained")
	}
	if state.Attributes["name"] != legacy.Attributes["name"] || state.Attributes["tag"] != legacy.Attributes["tag"] {
		t.Fatal("legacy baremetal state attributes changed")
	}
	if _, errors := validateTag("bad tag", "tag"); len(errors) == 0 {
		t.Fatal("invalid tag was accepted")
	}
	if _, errors := instance.Schema["boot_disk_size"].ValidateFunc(19, "boot_disk_size"); len(errors) == 0 {
		t.Fatal("undersized boot disk was accepted")
	}
}

func TestBareMetalServiceEmptyIDsRemainNotFound(t *testing.T) {
	client := &productClient{}
	for name, err := range map[string]error{
		"instance": clientEmptyInstanceError(client),
		"raid":     clientEmptyRaidError(client),
		"firewall": clientEmptyFirewallError(client),
	} {
		if err == nil || !isNotFoundError(err) {
			t.Errorf("%s empty ID error = %v, want provider not-found error", name, err)
		}
	}
}

func clientEmptyInstanceError(client *productClient) error {
	_, err := client.describeBareMetalInstanceById("")
	return err
}

func clientEmptyRaidError(client *productClient) error {
	_, err := client.getRaidTypeById("")
	return err
}

func clientEmptyFirewallError(client *productClient) error {
	_, err := client.describeFirewallById("")
	return err
}

func httpRequestHandlerCount(sdkClient interface{}) int {
	value := reflect.ValueOf(sdkClient)
	if value.Kind() != reflect.Ptr {
		return 0
	}
	value = value.Elem()
	if field := value.FieldByName("Client"); field.IsValid() {
		value = field.Elem()
	}
	field := value.FieldByName("httpRequestHandlers")
	if !field.IsValid() {
		return 0
	}
	return field.Len()
}

type testRuntime struct {
	config *ucloud.Config
	name   string
	calls  int
}

var _ product.RuntimeV1 = (*testRuntime)(nil)

func (runtime *testRuntime) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	if name != Name {
		return nil, fmt.Errorf("unexpected product %q", name)
	}
	runtime.name = name
	runtime.calls++
	return constructor(runtime.config, &auth.Credential{}, nil), nil
}

func assertSchemaFields(t *testing.T, resource *schema.Resource, want map[string]struct {
	typeValue schema.ValueType
	required  bool
	optional  bool
	computed  bool
	forceNew  bool
	sensitive bool
}) {
	t.Helper()
	if len(resource.Schema) != len(want) {
		t.Fatalf("schema has %d fields, want %d", len(resource.Schema), len(want))
	}
	for name, expectation := range want {
		field, ok := resource.Schema[name]
		if !ok {
			t.Fatalf("schema field %q is missing", name)
		}
		if field.Type != expectation.typeValue || field.Required != expectation.required || field.Optional != expectation.optional || field.Computed != expectation.computed || field.ForceNew != expectation.forceNew || field.Sensitive != expectation.sensitive {
			t.Errorf("schema field %q = type=%v required=%t optional=%t computed=%t forceNew=%t sensitive=%t", name, field.Type, field.Required, field.Optional, field.Computed, field.ForceNew, field.Sensitive)
		}
	}
}
