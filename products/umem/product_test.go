package umem

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestRegistrationKeepsLegacyTerraformSurface(t *testing.T) {
	provider := &schema.Provider{}
	if err := product.Register(provider, product.Bind(Name, New(), product.WithTerraformNamespaces("redis", "memcache"))); err != nil {
		t.Fatalf("register umem: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("validate umem provider: %v", err)
	}

	for _, name := range []string{"ucloud_redis_instance", "ucloud_memcache_instance"} {
		if provider.ResourcesMap[name] == nil {
			t.Errorf("resource %q is not registered", name)
		}
	}
	if len(provider.DataSourcesMap) != 0 {
		t.Fatalf("unexpected UMem data sources: %#v", provider.DataSourcesMap)
	}
}

func TestRegistrationUsesStableSchemaFields(t *testing.T) {
	redis := resourceUCloudRedisInstance()
	for _, field := range []string{
		"availability_zone", "standby_zone", "name", "instance_type", "engine_version",
		"charge_type", "duration", "vpc_id", "subnet_id", "password", "tag",
		"auto_backup", "backup_begin_time", "ip_set", "create_time", "expire_time", "status",
	} {
		if redis.Schema[field] == nil {
			t.Errorf("redis schema is missing field %q", field)
		}
	}
	if !redis.Schema["password"].Sensitive {
		t.Error("redis password schema must remain sensitive")
	}

	memcache := resourceUCloudMemcacheInstance()
	for _, field := range []string{
		"availability_zone", "name", "instance_type", "charge_type", "duration",
		"vpc_id", "subnet_id", "tag", "ip_set", "create_time", "expire_time", "status",
	} {
		if memcache.Schema[field] == nil {
			t.Errorf("memcache schema is missing field %q", field)
		}
	}
}

type runtimeStub struct {
	name  string
	calls int
}

var _ product.RuntimeV1 = (*runtimeStub)(nil)

func (stub *runtimeStub) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	stub.name = name
	stub.calls++
	config := ucloud.NewConfig()
	return constructor(&config, &auth.Credential{}, nil), nil
}

func TestClientFromMetaUsesProductRuntime(t *testing.T) {
	stub := &runtimeStub{}
	client, err := clientFromMeta(stub)
	if err != nil {
		t.Fatalf("get product client: %v", err)
	}
	if stub.name != Name {
		t.Fatalf("product client name = %q, want %q", stub.name, Name)
	}
	if stub.calls != 1 {
		t.Fatalf("product client calls = %d, want 1", stub.calls)
	}
	if client == nil || client.umemconn == nil || client.pumemconn == nil {
		t.Fatal("product client did not initialize public and private UMem SDK clients")
	}
}

func TestClientFromMetaRejectsInvalidRuntime(t *testing.T) {
	if _, err := clientFromMeta(struct{}{}); err == nil {
		t.Fatal("expected invalid runtime error")
	}
}

func TestInstanceTypeParsers(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		redis   bool
		memory  int
		wantErr bool
	}{
		{name: "redis master", value: "redis-master-2", redis: true, memory: 2},
		{name: "redis distributed", value: "redis-distributed-16", redis: true, memory: 16},
		{name: "memcache master", value: "memcache-master-4", memory: 4},
		{name: "invalid redis", value: "redis-standby-2", redis: true, wantErr: true},
		{name: "invalid memcache", value: "memcache-distributed-2", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.redis {
				parsed, err := parseRedisInstanceType(test.value)
				if test.wantErr {
					if err == nil {
						t.Fatal("expected parse error")
					}
					return
				}
				if err != nil || parsed.Memory != test.memory {
					t.Fatalf("parsed redis type = %#v, err = %v", parsed, err)
				}
				return
			}

			parsed, err := parseMemcacheInstanceType(test.value)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected parse error")
				}
				return
			}
			if err != nil || parsed.Memory != test.memory {
				t.Fatalf("parsed memcache type = %#v, err = %v", parsed, err)
			}
		})
	}
}
