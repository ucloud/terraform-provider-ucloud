package uhost

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestRegistrationKeepsLegacyTerraformSurface(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("instance", "instances", "images", "isolation_group"))); err != nil {
		t.Fatalf("register UHost product: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate UHost provider: %v", err)
	}

	for _, name := range []string{"ucloud_instance", "ucloud_instance_state", "ucloud_isolation_group"} {
		if provider.ResourcesMap[name] == nil {
			t.Errorf("resource %q is not registered", name)
		}
	}
	for _, name := range []string{"ucloud_instances", "ucloud_images"} {
		if provider.DataSourcesMap[name] == nil {
			t.Errorf("data source %q is not registered", name)
		}
	}
}

func TestInstanceSchemaCompatibility(t *testing.T) {
	instance := resourceUCloudInstance()
	assertResourceCallbacks(t, instance, "ucloud_instance")

	wantFields := map[string]schemaFieldExpectation{
		"availability_zone":          {typeValue: schema.TypeString, required: true, forceNew: true},
		"image_id":                   {typeValue: schema.TypeString, required: true},
		"root_password":              {typeValue: schema.TypeString, optional: true, computed: true, sensitive: true},
		"login_mode":                 {typeValue: schema.TypeString, optional: true},
		"key_pair_id":                {typeValue: schema.TypeString, optional: true},
		"deletion_protection":        {typeValue: schema.TypeBool, optional: true},
		"instance_type":              {typeValue: schema.TypeString, required: true},
		"name":                       {typeValue: schema.TypeString, optional: true, computed: true},
		"charge_type":                {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"duration":                   {typeValue: schema.TypeInt, optional: true, forceNew: true},
		"boot_disk_size":             {typeValue: schema.TypeInt, optional: true, computed: true},
		"boot_disk_type":             {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"data_disk_size":             {typeValue: schema.TypeInt, optional: true, computed: true},
		"data_disk_type":             {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"data_disks":                 {typeValue: schema.TypeList, optional: true},
		"network_interface":          {typeValue: schema.TypeList, optional: true},
		"delete_eips_with_instance":  {typeValue: schema.TypeBool, optional: true, forceNew: true},
		"delete_disks_with_instance": {typeValue: schema.TypeBool, optional: true, forceNew: true},
		"remark":                     {typeValue: schema.TypeString, optional: true, computed: true},
		"tag":                        {typeValue: schema.TypeString, optional: true},
		"security_group":             {typeValue: schema.TypeString, optional: true, computed: true},
		"security_mode":              {typeValue: schema.TypeString, optional: true, forceNew: true},
		"sec_group_id":               {typeValue: schema.TypeSet, optional: true},
		"isolation_group":            {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"vpc_id":                     {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"subnet_id":                  {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"private_ip":                 {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"allow_stopping_for_update":  {typeValue: schema.TypeBool, optional: true},
		"user_data":                  {typeValue: schema.TypeString, optional: true, forceNew: true},
		"min_cpu_platform":           {typeValue: schema.TypeString, optional: true, forceNew: true},
		"cpu_platform":               {typeValue: schema.TypeString, computed: true},
		"cpu":                        {typeValue: schema.TypeInt, computed: true},
		"memory":                     {typeValue: schema.TypeInt, computed: true},
		"status":                     {typeValue: schema.TypeString, computed: true},
		"disk_set":                   {typeValue: schema.TypeList, computed: true},
		"ip_set":                     {typeValue: schema.TypeList, computed: true},
		"create_time":                {typeValue: schema.TypeString, computed: true},
		"expire_time":                {typeValue: schema.TypeString, computed: true},
		"auto_renew":                 {typeValue: schema.TypeBool, computed: true},
		"rdma_cluster_id":            {typeValue: schema.TypeString, computed: true},
	}
	assertSchemaFields(t, instance, wantFields)

	if instance.Schema["tag"].Default != defaultTag {
		t.Fatalf("tag default = %#v, want %q", instance.Schema["tag"].Default, defaultTag)
	}
	if instance.Schema["min_cpu_platform"].Default != "Intel/Auto" {
		t.Fatalf("min_cpu_platform default = %#v, want Intel/Auto", instance.Schema["min_cpu_platform"].Default)
	}
	if instance.Schema["data_disks"].MinItems != 1 || instance.Schema["data_disks"].MaxItems != 1 {
		t.Fatalf("data_disks cardinality = %d..%d, want 1..1", instance.Schema["data_disks"].MinItems, instance.Schema["data_disks"].MaxItems)
	}
	if instance.Schema["network_interface"].MinItems != 1 || instance.Schema["network_interface"].MaxItems != 1 {
		t.Fatalf("network_interface cardinality = %d..%d, want 1..1", instance.Schema["network_interface"].MinItems, instance.Schema["network_interface"].MaxItems)
	}
	if instance.Schema["sec_group_id"].MaxItems != 5 {
		t.Fatalf("sec_group_id MaxItems = %d, want 5", instance.Schema["sec_group_id"].MaxItems)
	}
}

func TestInstanceStateAndIsolationGroupCompatibility(t *testing.T) {
	state := resourceUCloudInstanceState()
	if state.Create == nil || state.Read == nil || state.Update == nil || state.Delete == nil {
		t.Fatal("ucloud_instance_state CRUD callbacks are incomplete")
	}
	if state.Importer == nil || state.Importer.State == nil {
		t.Fatal("ucloud_instance_state importer is missing")
	}
	assertSchemaFields(t, state, map[string]schemaFieldExpectation{
		"instance_id": {typeValue: schema.TypeString, required: true},
		"state":       {typeValue: schema.TypeString, required: true},
		"force":       {typeValue: schema.TypeBool, optional: true},
	})

	isolationGroup := resourceUCloudIsolationGroup()
	if isolationGroup.Create == nil || isolationGroup.Read == nil || isolationGroup.Delete == nil {
		t.Fatal("ucloud_isolation_group CRUD callbacks are incomplete")
	}
	if isolationGroup.Importer == nil || isolationGroup.Importer.State == nil {
		t.Fatal("ucloud_isolation_group importer is missing")
	}
	assertSchemaFields(t, isolationGroup, map[string]schemaFieldExpectation{
		"name":   {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
		"remark": {typeValue: schema.TypeString, optional: true, computed: true, forceNew: true},
	})
}

func TestDataSourcesPreserveSchemaAndStateMigration(t *testing.T) {
	images := dataSourceUCloudImages()
	assertDataSourceFields(t, images, []string{"availability_zone", "name_regex", "most_recent", "image_type", "os_type", "image_id", "ids", "output_file", "total_count", "images"})
	imageFields, ok := images.Schema["images"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("images element type = %T, want *schema.Resource", images.Schema["images"].Elem)
	}
	for _, field := range []string{"id", "name", "type", "size", "availability_zone", "os_type", "os_name", "features", "create_time", "description", "status"} {
		if imageFields.Schema[field] == nil || !imageFields.Schema[field].Computed {
			t.Errorf("images nested field %q is not computed", field)
		}
	}

	instances := dataSourceUCloudInstances()
	assertDataSourceFields(t, instances, []string{"availability_zone", "name_regex", "ids", "tag", "output_file", "total_count", "instances"})
	if instances.SchemaVersion != 1 || instances.MigrateState == nil {
		t.Fatalf("instances state migration = version %d callback %v, want version 1 with callback", instances.SchemaVersion, instances.MigrateState != nil)
	}
	instanceFields, ok := instances.Schema["instances"].Elem.(*schema.Resource)
	if !ok {
		t.Fatalf("instances element type = %T, want *schema.Resource", instances.Schema["instances"].Elem)
	}
	for _, field := range []string{"availability_zone", "id", "name", "cpu", "memory", "instance_type", "charge_type", "auto_renew", "remark", "tag", "status", "vpc_id", "subnet_id", "private_ip", "create_time", "expire_time", "disk_set", "ip_set"} {
		if instanceFields.Schema[field] == nil || !instanceFields.Schema[field].Computed {
			t.Errorf("instances nested field %q is not computed", field)
		}
	}

	state := &terraform.InstanceState{
		ID: "foo",
		Attributes: map[string]string{
			"instances.0.auto_renew":         "Yes",
			"instances.0.disk_set.0.is_boot": "No",
			"instances.0.disk_set.0.type":    "CLOUD_SSD",
			"instances.0.memory":             "2048",
			"instances.0.charge_type":        "Dynamic",
			"instances.0.name":               "unchanged",
		},
	}
	want := &terraform.InstanceState{
		ID: "foo",
		Attributes: map[string]string{
			"instances.0.auto_renew":         "true",
			"instances.0.disk_set.0.is_boot": "false",
			"instances.0.disk_set.0.type":    "cloud_ssd",
			"instances.0.memory":             "2",
			"instances.0.charge_type":        "dynamic",
			"instances.0.name":               "unchanged",
		},
	}
	got, err := dataSourceUCloudInstancesMigrateState(0, state, nil)
	if err != nil {
		t.Fatalf("migrate instances state: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("migrated instances state = %#v, want %#v", got, want)
	}
}

func TestClientComposesLegacyClientsAndTimeouts(t *testing.T) {
	config := &ucloud.Config{Region: "cn-bj2", ProjectId: "org-test", Timeout: time.Second}
	credential := &auth.Credential{}
	rawClient := newClient(config, credential, nil)
	client, ok := rawClient.(*productClient)
	if !ok {
		t.Fatalf("newClient() returned %T, want *productClient", rawClient)
	}
	if client.uhostconn == nil || client.unetconn == nil || client.vpcconn == nil {
		t.Fatal("UHost client did not compose exactly the required SDK clients")
	}
	if got := client.uhostconn.GetConfig().Timeout; got != 60*time.Second {
		t.Fatalf("UHost timeout = %s, want %s", got, 60*time.Second)
	}
	for name, sdkConfig := range map[string]*ucloud.Config{
		"UHost": client.uhostconn.GetConfig(),
		"UNet":  client.unetconn.GetConfig(),
		"VPC":   client.vpcconn.GetConfig(),
	} {
		if sdkConfig.Region != config.Region || sdkConfig.ProjectId != config.ProjectId {
			t.Errorf("%s config = region %q project %q, want region %q project %q", name, sdkConfig.Region, sdkConfig.ProjectId, config.Region, config.ProjectId)
		}
	}
	if got := client.unetconn.GetConfig().Timeout; got != config.Timeout {
		t.Fatalf("UNet timeout = %s, want caller timeout %s", got, config.Timeout)
	}
	if got := client.vpcconn.GetConfig().Timeout; got != config.Timeout {
		t.Fatalf("VPC timeout = %s, want caller timeout %s", got, config.Timeout)
	}
	if config.Timeout != time.Second {
		t.Fatalf("newClient changed caller timeout to %s", config.Timeout)
	}
}

func TestClientFromMetaUsesProductRuntime(t *testing.T) {
	runtime := &testRuntime{config: &ucloud.Config{Region: "cn-bj2", ProjectId: "org-test"}}
	client, err := clientFromMeta(runtime)
	if err != nil {
		t.Fatalf("clientFromMeta() error = %v", err)
	}
	if runtime.name != Name || runtime.calls != 1 {
		t.Fatalf("runtime calls = name %q calls %d, want name %q calls 1", runtime.name, runtime.calls, Name)
	}
	if client == nil {
		t.Fatal("clientFromMeta() returned nil client")
	}
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("clientFromMeta accepted a non-runtime value")
	}
}

type schemaFieldExpectation struct {
	typeValue schema.ValueType
	required  bool
	optional  bool
	computed  bool
	forceNew  bool
	sensitive bool
}

func assertResourceCallbacks(t *testing.T, resource *schema.Resource, name string) {
	t.Helper()
	if resource == nil {
		t.Fatalf("%s is nil", name)
	}
	if resource.Create == nil || resource.Read == nil || resource.Update == nil || resource.Delete == nil {
		t.Fatalf("%s CRUD callbacks are incomplete", name)
	}
	if resource.CustomizeDiff == nil {
		t.Fatalf("%s CustomizeDiff is missing", name)
	}
	if resource.Importer == nil || resource.Importer.State == nil {
		t.Fatalf("%s importer is missing", name)
	}
	if resource.Timeouts == nil || resource.Timeouts.Create == nil || resource.Timeouts.Update == nil || resource.Timeouts.Delete == nil {
		t.Fatalf("%s timeouts are incomplete", name)
	}
	want := map[string]time.Duration{"create": 30 * time.Minute, "update": 20 * time.Minute, "delete": 15 * time.Minute}
	got := map[string]*time.Duration{
		"create": resource.Timeouts.Create,
		"update": resource.Timeouts.Update,
		"delete": resource.Timeouts.Delete,
	}
	for timeoutName, wantDuration := range want {
		if got[timeoutName] == nil || *got[timeoutName] != wantDuration {
			t.Errorf("%s timeout = %v, want %s", timeoutName, got[timeoutName], wantDuration)
		}
	}
}

func assertSchemaFields(t *testing.T, resource *schema.Resource, want map[string]schemaFieldExpectation) {
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

func assertDataSourceFields(t *testing.T, dataSource *schema.Resource, want []string) {
	t.Helper()
	if dataSource == nil || dataSource.Read == nil {
		t.Fatal("data source or Read callback is missing")
	}
	if len(dataSource.Schema) != len(want) {
		t.Fatalf("data source schema has %d fields, want %d", len(dataSource.Schema), len(want))
	}
	for _, name := range want {
		if dataSource.Schema[name] == nil {
			t.Errorf("data source field %q is missing", name)
		}
	}
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
