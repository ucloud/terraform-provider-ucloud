package umem

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-ucloud/internal/product"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// These tests cover the CRUD unit level required by spec 10.1 for every
// resource: request conversion, response expansion, NotFound handling and error
// propagation. They talk to an httptest server that speaks the UCloud API
// envelope instead of the cloud, so no remote resource is created.
//
// The provider's own HTTP layer is exercised end to end (resource function ->
// productClient -> SDK client -> HTTP), which is what keeps the base64 password
// encoding, the zone parameter of the UMemSpace family and the per-shard resize
// loop honest.

// umemAPIStub records the requests made by the resource implementations and
// answers them from a per-action script.
type umemAPIStub struct {
	t       *testing.T
	mu      sync.Mutex
	forms   []url.Values
	scripts map[string]func(url.Values) string
}

func newUmemAPIStub(t *testing.T) *umemAPIStub {
	return &umemAPIStub{t: t, scripts: map[string]func(url.Values) string{}}
}

// on scripts a fixed response body for an action.
func (stub *umemAPIStub) on(action, body string) *umemAPIStub {
	stub.scripts[action] = func(url.Values) string { return body }
	return stub
}

// onFunc scripts a response body computed from the request, which also allows
// stateful answers (e.g. a shard that reports its old size before the resize).
func (stub *umemAPIStub) onFunc(action string, body func(url.Values) string) *umemAPIStub {
	stub.scripts[action] = body
	return stub
}

func (stub *umemAPIStub) start() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		raw, err := io.ReadAll(request.Body)
		if err != nil {
			stub.t.Errorf("read request body: %v", err)
		}
		// These tests must never reach the cloud. The resource code posts to
		// config.BaseUrl, which the test runtime points at this stub; fail loudly
		// if a future change ever sends a request anywhere but loopback.
		if host, _, splitErr := net.SplitHostPort(request.Host); splitErr != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
			stub.t.Errorf("UMem API request went to non-loopback host %q: unit tests must not touch the cloud", request.Host)
		}
		form, err := url.ParseQuery(string(raw))
		if err != nil {
			stub.t.Errorf("parse request form: %v", err)
		}

		stub.mu.Lock()
		stub.forms = append(stub.forms, form)
		script := stub.scripts[form.Get("Action")]
		stub.mu.Unlock()

		writer.Header().Set("Content-Type", "application/json")
		if script == nil {
			// Fail loudly: an unscripted call usually means the resource grew a
			// new API call that no test covers yet.
			stub.t.Errorf("unexpected UMem API action %q with params %v", form.Get("Action"), form)
			_, _ = fmt.Fprint(writer, `{"RetCode":0}`)
			return
		}
		_, _ = fmt.Fprint(writer, script(form))
	}))
}

// calls returns the recorded requests for one action, in order.
func (stub *umemAPIStub) calls(action string) []url.Values {
	stub.mu.Lock()
	defer stub.mu.Unlock()

	var result []url.Values
	for _, form := range stub.forms {
		if form.Get("Action") == action {
			result = append(result, form)
		}
	}
	return result
}

func (stub *umemAPIStub) count(action string) int {
	return len(stub.calls(action))
}

// only returns the single recorded request for an action and fails the test when
// the action was called zero or multiple times.
func (stub *umemAPIStub) only(t *testing.T, action string) url.Values {
	t.Helper()

	calls := stub.calls(action)
	if len(calls) != 1 {
		t.Fatalf("%s was called %d times, want exactly 1 (params: %v)", action, len(calls), stub.actions())
	}
	return calls[0]
}

func (stub *umemAPIStub) actions() []string {
	stub.mu.Lock()
	defer stub.mu.Unlock()

	names := make([]string, 0, len(stub.forms))
	for _, form := range stub.forms {
		names = append(names, form.Get("Action"))
	}
	return names
}

// umemTestRuntime is the product.RuntimeV1 the resources see in these tests. It
// hands the resource the same productClient the provider builds, but points the
// SDK clients at the stub server.
type umemTestRuntime struct {
	config     ucloud.Config
	credential auth.Credential
	calls      int
}

var _ product.RuntimeV1 = (*umemTestRuntime)(nil)

func (runtime *umemTestRuntime) ProductClient(name string, constructor product.ClientConstructor) (interface{}, error) {
	runtime.calls++
	return constructor(&runtime.config, &runtime.credential, nil), nil
}

func newUmemTestRuntime(serverURL string) *umemTestRuntime {
	config := ucloud.NewConfig()
	config.BaseUrl = serverURL
	config.Region = "cn-sh2"
	config.ProjectId = "org-test"

	return &umemTestRuntime{config: config}
}

// newUmemTestClient builds the product client the same way the runtime does, for
// the tests that call an update helper directly.
func newUmemTestClient(t *testing.T, serverURL string) *productClient {
	t.Helper()

	client, err := clientFromMeta(newUmemTestRuntime(serverURL))
	if err != nil {
		t.Fatalf("build product client: %v", err)
	}
	return client
}

// newUmemTestData builds resource data without a diff. HasChange is false for
// every attribute, which is exactly the "create from configuration" view of the
// resource the create path works with, and it keeps the trailing update pass
// from firing unrelated branches.
func newUmemTestData(t *testing.T, resource *schema.Resource, raw map[string]interface{}) *schema.ResourceData {
	t.Helper()

	data := resource.Data(nil)
	for key, value := range raw {
		if err := data.Set(key, value); err != nil {
			t.Fatalf("set %s: %v", key, err)
		}
	}
	return data
}

func redisInstanceListBody(size int, state string) string {
	return fmt.Sprintf(`{"RetCode":0,"TotalCount":1,"DataSet":[{
		"Zone":"cn-sh2-01","GroupId":"uredis-1","Name":"tf-acc-redis","Tag":"tf-acc",
		"VPCId":"uvnet-1","SubnetId":"subnet-1","Size":%d,"MemorySize":%d,"State":%q,
		"VirtualIP":"10.0.0.7","Port":6379,"Version":"4.0","SlaveZone":"cn-sh2-02",
		"AutoBackup":"disable","BackupTime":3,"ChargeType":"Month",
		"CreateTime":1700000000,"ExpireTime":0}]}`, size, size, state)
}

func distributedRedisListBody(size int, state string) string {
	return fmt.Sprintf(`{"RetCode":0,"TotalCount":1,"DataSet":[{
		"Zone":"cn-sh2-01","SpaceId":"udredis-1","Name":"tf-acc-redis","Tag":"tf-acc",
		"VPCId":"uvnet-1","SubnetId":"subnet-1","Size":%d,"State":%q,"ChargeType":"Month",
		"CreateTime":1700000000,"ExpireTime":0,
		"Address":[{"IP":"10.0.0.8","Port":6379}]}]}`, size, state)
}

func blockInfoBody(readMode string, blocks ...string) string {
	return fmt.Sprintf(`{"RetCode":0,"ReadMode":%q,"TotalCount":%d,"DataSet":[%s]}`,
		readMode, len(blocks), strings.Join(blocks, ","))
}

func block(blockID string, size int, state string) string {
	return fmt.Sprintf(`{"BlockId":%q,"BlockName":%q,"BlockPort":6379,"BlockReadWeight":0,
		"BlockSize":%d,"BlockSlotBegin":0,"BlockSlotEnd":8191,"BlockState":%q,"BlockType":"master",
		"BlockUsedSize":0,"BlockVip":"10.0.0.9"}`, blockID, blockID+"-name", size, state)
}

func proxyInfoBody(proxies ...string) string {
	return fmt.Sprintf(`{"RetCode":0,"TotalCount":%d,"DataSet":[%s]}`, len(proxies), strings.Join(proxies, ","))
}

// scriptActiveStandbyCreate scripts every call the active-standby create path
// makes, including the trailing read.
func scriptActiveStandbyCreate(stub *umemAPIStub, size int) {
	stub.on("DescribeURedisConfig", `{"RetCode":0,"TotalCount":1,"DataSet":[
		{"Zone":"cn-sh2-01","ConfigId":"cfg-default","Version":"4.0","IsModify":"Unmodifiable","State":"Usable"}]}`)
	stub.on("CreateURedisGroup", `{"RetCode":0,"GroupId":"uredis-1"}`)
	stub.on("DescribeURedisGroup", redisInstanceListBody(size, statusRunning))
	stub.on("ModifyURedisGroupName", `{"RetCode":0}`)
	stub.on("ResizeURedisGroup", `{"RetCode":0}`)
	stub.on("ModifyURedisGroupPassword", `{"RetCode":0}`)
	stub.on("UpdateURedisBackupStrategy", `{"RetCode":0}`)
}

func scriptDistributedCreate(stub *umemAPIStub, size, shardCount int) {
	blocks := make([]string, 0, shardCount)
	for index := 0; index < shardCount; index++ {
		blocks = append(blocks, block(fmt.Sprintf("udredis-1_%d", index), size/shardCount, statusRunning))
	}
	stub.on("CreateUMemSpace", `{"RetCode":0,"SpaceId":"udredis-1"}`)
	stub.on("DescribeUMemSpace", distributedRedisListBody(size, statusRunning))
	stub.on("DescribeUMemBlockInfo", blockInfoBody("master", blocks...))
	stub.on("DescribeUDRedisProxyInfo", proxyInfoBody())
	stub.on("ModifyUMemSpaceName", `{"RetCode":0}`)
	stub.on("ModifyUMemPassword", `{"RetCode":0}`)
	stub.on("ResizeUDRedisBlockSize", `{"RetCode":0}`)
}

// TestRedisInstanceCreateConvertsRequestAndExpandsResponse covers the create
// request conversion (VPC/subnet, zone, standby zone, size, parameter group and
// the encoded password) and the response expansion into state.
func TestRedisInstanceCreateConvertsRequestAndExpandsResponse(t *testing.T) {
	stub := newUmemAPIStub(t)
	scriptActiveStandbyCreate(stub, 2)
	server := stub.start()
	defer server.Close()

	data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"standby_zone":      "cn-sh2-02",
		"name":              "tf-acc-redis",
		"instance_type":     "redis-master-2",
		"engine_version":    "4.0",
		"charge_type":       "month",
		"duration":          1,
		"password":          "2018_tfacc",
		"tag":               "tf-acc",
		"vpc_id":            "uvnet-1",
		"subnet_id":         "subnet-1",
	})
	data.SetId("")

	if err := resourceUCloudRedisInstanceCreate(data, newUmemTestRuntime(server.URL)); err != nil {
		t.Fatalf("create active-standby redis: %v", err)
	}

	create := stub.only(t, "CreateURedisGroup")
	for field, want := range map[string]string{
		"Zone":             "cn-sh2-01",
		"SlaveZone":        "cn-sh2-02",
		"VPCId":            "uvnet-1",
		"SubnetId":         "subnet-1",
		"Size":             "2",
		"HighAvailability": "enable",
		"ConfigId":         "cfg-default",
		"ChargeType":       "Month",
		"Quantity":         "1",
		"Name":             "tf-acc-redis",
		"Version":          "4.0",
		"AutoBackup":       "disable",
		"BackupTime":       "3",
	} {
		if got := create.Get(field); got != want {
			t.Errorf("CreateURedisGroup %s = %q, want %q", field, got, want)
		}
	}
	// The private SDK base64-encodes the password for CreateURedisGroup.
	if got, want := create.Get("Password"), base64.StdEncoding.EncodeToString([]byte("2018_tfacc")); got != want {
		t.Errorf("CreateURedisGroup Password = %q, want %q", got, want)
	}

	if got := data.Id(); got != "uredis-1" {
		t.Errorf("resource ID = %q, want uredis-1", got)
	}
	for field, want := range map[string]interface{}{
		"vpc_id":            "uvnet-1",
		"subnet_id":         "subnet-1",
		"status":            statusRunning,
		"instance_type":     "redis-master-2",
		"availability_zone": "cn-sh2-01",
		"standby_zone":      "cn-sh2-02",
		"tag":               "tf-acc",
	} {
		if got := data.Get(field); got != want {
			t.Errorf("state %s = %#v, want %#v", field, got, want)
		}
	}
}

// TestRedisInstanceCreateRejectsMissingGroupID covers the guard added for the
// acceptance failure where CreateURedisGroup answered without a GroupId: an
// empty id must never reach state.
func TestRedisInstanceCreateRejectsMissingGroupID(t *testing.T) {
	for name, body := range map[string]string{
		"no group id": `{"RetCode":0}`,
		"empty id":    `{"RetCode":0,"GroupId":""}`,
		"empty body":  ``,
	} {
		t.Run(name, func(t *testing.T) {
			stub := newUmemAPIStub(t)
			scriptActiveStandbyCreate(stub, 2)
			stub.on("CreateURedisGroup", body)
			server := stub.start()
			defer server.Close()

			data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
				"availability_zone": "cn-sh2-01",
				"name":              "tf-acc-redis",
				"instance_type":     "redis-master-2",
				"engine_version":    "4.0",
				"vpc_id":            "uvnet-1",
				"subnet_id":         "subnet-1",
			})

			if err := resourceUCloudRedisInstanceCreate(data, newUmemTestRuntime(server.URL)); err == nil {
				t.Fatal("expected create to fail when the response carries no instance id")
			}
			if got := data.Id(); got != "" {
				t.Errorf("resource ID after failed create = %q, want empty", got)
			}
			if got := stub.count("DescribeURedisGroup"); got != 0 {
				t.Errorf("created an empty id and polled the API %d times, want 0", got)
			}
		})
	}
}

// TestDistributedRedisCreateConvertsRequestAndExpandsResponse covers the
// distributed create path: the shard count default, protocol, VPC/subnet and the
// block/proxy/read_mode expansion of the trailing read.
func TestDistributedRedisCreateConvertsRequestAndExpandsResponse(t *testing.T) {
	stub := newUmemAPIStub(t)
	scriptDistributedCreate(stub, 16, defaultDistributedRedisBlockCnt)
	server := stub.start()
	defer server.Close()

	data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
		"availability_zone": "cn-sh2-01",
		"name":              "tf-acc-redis",
		"instance_type":     "redis-distributed-16",
		"charge_type":       "month",
		"duration":          1,
		"password":          "2018_tfacc",
		"vpc_id":            "uvnet-1",
		"subnet_id":         "subnet-1",
	})

	if err := resourceUCloudRedisInstanceCreate(data, newUmemTestRuntime(server.URL)); err != nil {
		t.Fatalf("create distributed redis: %v", err)
	}

	create := stub.only(t, "CreateUMemSpace")
	for field, want := range map[string]string{
		"Zone":     "cn-sh2-01",
		"VPCId":    "uvnet-1",
		"SubnetId": "subnet-1",
		"Size":     "16",
		"Protocol": "redis",
		"BlockCnt": "4",
		"Quantity": "1",
	} {
		if got := create.Get(field); got != want {
			t.Errorf("CreateUMemSpace %s = %q, want %q", field, got, want)
		}
	}
	// The public SDK base64-encodes the password for CreateUMemSpace.
	if got, want := create.Get("Password"), base64.StdEncoding.EncodeToString([]byte("2018_tfacc")); got != want {
		t.Errorf("CreateUMemSpace Password = %q, want %q", got, want)
	}
	if got := data.Id(); got != "udredis-1" {
		t.Errorf("resource ID = %q, want udredis-1", got)
	}
	if got := data.Get("block_cnt").(int); got != defaultDistributedRedisBlockCnt {
		t.Errorf("state block_cnt = %d, want %d", got, defaultDistributedRedisBlockCnt)
	}
	if got := data.Get("read_mode"); got != "master" {
		t.Errorf("state read_mode = %#v, want master", got)
	}
}

func TestDistributedRedisCreateRejectsMissingSpaceID(t *testing.T) {
	for name, body := range map[string]string{
		"no space id": `{"RetCode":0}`,
		"empty id":    `{"RetCode":0,"SpaceId":""}`,
		"empty body":  ``,
	} {
		t.Run(name, func(t *testing.T) {
			stub := newUmemAPIStub(t)
			scriptDistributedCreate(stub, 16, defaultDistributedRedisBlockCnt)
			stub.on("CreateUMemSpace", body)
			server := stub.start()
			defer server.Close()

			data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
				"availability_zone": "cn-sh2-01",
				"name":              "tf-acc-redis",
				"instance_type":     "redis-distributed-16",
				"charge_type":       "month",
				"vpc_id":            "uvnet-1",
				"subnet_id":         "subnet-1",
			})

			if err := resourceUCloudRedisInstanceCreate(data, newUmemTestRuntime(server.URL)); err == nil {
				t.Fatal("expected create to fail when the response carries no space id")
			}
			if got := data.Id(); got != "" {
				t.Errorf("resource ID after failed create = %q, want empty", got)
			}
			if got := stub.count("DescribeUMemSpace"); got != 0 {
				t.Errorf("created an empty id and polled the API %d times, want 0", got)
			}
		})
	}
}

// TestRedisInstanceUpdateConvertsPasswordAndRestartRequests covers the update
// helpers added by the umem 1.1 work. Both the base64 password encoding and the
// "restart requires Running" pre-check are request-conversion details of the
// update path.
func TestRedisInstanceUpdateConvertsPasswordAndRestartRequests(t *testing.T) {
	t.Run("password is base64 encoded", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("ModifyURedisGroupPassword", `{"RetCode":0}`)
		stub.on("DescribeURedisGroup", redisInstanceListBody(2, statusRunning))
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
			"instance_type": "redis-master-2",
			"password":      "2019_tfacc",
		})
		data.SetId("uredis-1")

		if err := modifyActiveStandbyRedisPassword(newUmemTestClient(t, server.URL), data); err != nil {
			t.Fatalf("modify password: %v", err)
		}

		request := stub.only(t, "ModifyURedisGroupPassword")
		if got := request.Get("GroupId"); got != "uredis-1" {
			t.Errorf("GroupId = %q, want uredis-1", got)
		}
		if got, want := request.Get("Password"), base64.StdEncoding.EncodeToString([]byte("2019_tfacc")); got != want {
			t.Errorf("Password = %q, want %q", got, want)
		}
	})

	t.Run("restart requires a running instance", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("RestartURedisGroup", `{"RetCode":0}`)
		stub.on("DescribeURedisGroup", redisInstanceListBody(2, statusUNBind))
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
			"instance_type": "redis-master-2",
		})
		data.SetId("uredis-1")

		err := restartActiveStandbyRedis(newUmemTestClient(t, server.URL), data)
		if err == nil || !strings.Contains(err.Error(), "must be Running") {
			t.Fatalf("restart error = %v, want the isolated-state rejection", err)
		}
		if got := stub.count("RestartURedisGroup"); got != 0 {
			t.Errorf("RestartURedisGroup was called %d times for an isolated instance, want 0", got)
		}
	})

	t.Run("restart is sent for a running instance", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("RestartURedisGroup", `{"RetCode":0}`)
		stub.on("DescribeURedisGroup", redisInstanceListBody(2, statusRunning))
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
			"instance_type": "redis-master-2",
		})
		data.SetId("uredis-1")

		if err := restartActiveStandbyRedis(newUmemTestClient(t, server.URL), data); err != nil {
			t.Fatalf("restart: %v", err)
		}
		if got := stub.only(t, "RestartURedisGroup").Get("GroupId"); got != "uredis-1" {
			t.Errorf("RestartURedisGroup GroupId = %q, want uredis-1", got)
		}
	})
}

// TestRedisInstanceTransformRequest covers ISolationURedisGroup request
// conversion, including the empty-value guard: transform_type is not read back
// from the API, so removing it must fail with a clear message instead of sending
// an empty transform type.
func TestRedisInstanceTransformRequest(t *testing.T) {
	t.Run("unbind is sent for isolation", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("ISolationURedisGroup", `{"RetCode":0}`)
		stub.on("DescribeURedisGroup", redisInstanceListBody(2, statusUNBind))
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
			"instance_type":  "redis-master-2",
			"transform_type": "UNBind",
		})
		data.SetId("uredis-1")

		if err := transformActiveStandbyRedis(newUmemTestClient(t, server.URL), data); err != nil {
			t.Fatalf("isolate instance: %v", err)
		}
		request := stub.only(t, "ISolationURedisGroup")
		if got := request.Get("GroupId"); got != "uredis-1" {
			t.Errorf("GroupId = %q, want uredis-1", got)
		}
		if got := request.Get("TransformType"); got != "UNBind" {
			t.Errorf("TransformType = %q, want UNBind", got)
		}
	})

	t.Run("empty transform type is rejected without an API call", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
			"instance_type": "redis-master-2",
		})
		data.SetId("uredis-1")

		err := transformActiveStandbyRedis(newUmemTestClient(t, server.URL), data)
		if err == nil || !strings.Contains(err.Error(), "transform_type") {
			t.Fatalf("transform error = %v, want an explicit transform_type message", err)
		}
		if got := stub.count("ISolationURedisGroup"); got != 0 {
			t.Errorf("ISolationURedisGroup was called %d times for an empty transform type, want 0", got)
		}
	})
}

// TestDistributedRedisResizeCallsResizeUDRedisBlockSizePerShard covers the
// distributed resize introduced by umem 1.1: ResizeUMemSpace must not be used,
// every shard is resized through ResizeUDRedisBlockSize, shards already at the
// target size are skipped and invalid requests fail before any API call.
func TestDistributedRedisResizeCallsResizeUDRedisBlockSizePerShard(t *testing.T) {
	t.Run("each shard is resized to the target size", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		var blockCalls int
		stub.onFunc("DescribeUMemBlockInfo", func(url.Values) string {
			blockCalls++
			if blockCalls == 1 {
				// the shards still report the old per-shard size
				return blockInfoBody("master", block("udredis-1_0", 4, statusRunning), block("udredis-1_1", 4, statusRunning))
			}
			return blockInfoBody("master", block("udredis-1_0", 8, statusRunning), block("udredis-1_1", 8, statusRunning))
		})
		stub.on("ResizeUDRedisBlockSize", `{"RetCode":0}`)
		stub.on("DescribeUMemSpace", distributedRedisListBody(16, statusRunning))
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
			"availability_zone": "cn-sh2-01",
			"instance_type":     "redis-distributed-16",
		})
		data.SetId("udredis-1")

		if err := resizeDistributedRedisInstance(newUmemTestClient(t, server.URL), data); err != nil {
			t.Fatalf("resize distributed redis: %v", err)
		}

		resizes := stub.calls("ResizeUDRedisBlockSize")
		if len(resizes) != 2 {
			t.Fatalf("ResizeUDRedisBlockSize called %d times, want 2 (params: %v)", len(resizes), stub.actions())
		}
		for index, wantBlock := range []string{"udredis-1_0", "udredis-1_1"} {
			request := resizes[index]
			if got := request.Get("BlockId"); got != wantBlock {
				t.Errorf("resize %d BlockId = %q, want %q", index, got, wantBlock)
			}
			if got := request.Get("BlockSize"); got != "8" {
				t.Errorf("resize %d BlockSize = %q, want 8", index, got)
			}
			if got := request.Get("SpaceId"); got != "udredis-1" {
				t.Errorf("resize %d SpaceId = %q, want udredis-1", index, got)
			}
			if got := request.Get("Zone"); got != "cn-sh2-01" {
				t.Errorf("resize %d Zone = %q, want cn-sh2-01", index, got)
			}
		}
	})

	t.Run("shards already at the target size are skipped", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("DescribeUMemBlockInfo", blockInfoBody("master",
			block("udredis-1_0", 8, statusRunning), block("udredis-1_1", 8, statusRunning)))
		stub.on("ResizeUDRedisBlockSize", `{"RetCode":0}`)
		stub.on("DescribeUMemSpace", distributedRedisListBody(16, statusRunning))
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
			"availability_zone": "cn-sh2-01",
			"instance_type":     "redis-distributed-16",
		})
		data.SetId("udredis-1")

		if err := resizeDistributedRedisInstance(newUmemTestClient(t, server.URL), data); err != nil {
			t.Fatalf("resize distributed redis: %v", err)
		}
		if got := stub.count("ResizeUDRedisBlockSize"); got != 0 {
			t.Errorf("ResizeUDRedisBlockSize called %d times for shards already at target, want 0", got)
		}
		if got := stub.count("ResizeUMemSpace"); got != 0 {
			t.Errorf("ResizeUMemSpace called %d times, want 0 (the backend accepts it but never resizes)", got)
		}
	})

	for name, test := range map[string]struct {
		instanceType string
		blocks       []string
		wantMessage  string
	}{
		"total size not divisible by shard count": {
			instanceType: "redis-distributed-20",
			blocks: []string{
				block("udredis-1_0", 8, statusRunning),
				block("udredis-1_1", 8, statusRunning),
				block("udredis-1_2", 8, statusRunning),
			},
			wantMessage: "20GB is not divisible by its 3 shards",
		},
		"per-shard size outside the allowed set": {
			instanceType: "redis-distributed-20",
			blocks:       []string{block("udredis-1_0", 4, statusRunning), block("udredis-1_1", 4, statusRunning)},
			wantMessage:  "per-shard size 10GB is invalid",
		},
		"no shards reported": {
			instanceType: "redis-distributed-16",
			blocks:       nil,
			wantMessage:  "no shard information",
		},
	} {
		t.Run(name, func(t *testing.T) {
			stub := newUmemAPIStub(t)
			stub.on("DescribeUMemBlockInfo", blockInfoBody("master", test.blocks...))
			stub.on("ResizeUDRedisBlockSize", `{"RetCode":0}`)
			server := stub.start()
			defer server.Close()

			data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
				"availability_zone": "cn-sh2-01",
				"instance_type":     test.instanceType,
			})
			data.SetId("udredis-1")

			err := resizeDistributedRedisInstance(newUmemTestClient(t, server.URL), data)
			if err == nil || !strings.Contains(err.Error(), test.wantMessage) {
				t.Fatalf("resize error = %v, want message containing %q", err, test.wantMessage)
			}
			if got := stub.count("ResizeUDRedisBlockSize"); got != 0 {
				t.Errorf("ResizeUDRedisBlockSize called %d times for an invalid resize, want 0", got)
			}
		})
	}
}

// TestRedisInstanceReadHandlesNotFoundAndAPIErrors covers the two read outcomes
// spec 10.1 asks for: a missing instance clears the ID, and an API error
// propagates with the resource ID in the message.
func TestRedisInstanceReadHandlesNotFoundAndAPIErrors(t *testing.T) {
	t.Run("active-standby instance missing clears the ID", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("DescribeURedisGroup", `{"RetCode":0,"TotalCount":0,"DataSet":[]}`)
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
			"instance_type": "redis-master-2",
		})
		data.SetId("uredis-1")

		if err := resourceUCloudRedisInstanceRead(data, newUmemTestRuntime(server.URL)); err != nil {
			t.Fatalf("read missing instance: %v", err)
		}
		if got := data.Id(); got != "" {
			t.Errorf("resource ID = %q, want empty for a missing instance", got)
		}
	})

	t.Run("distributed instance missing clears the ID", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("DescribeUMemSpace", `{"RetCode":0,"TotalCount":0,"DataSet":[]}`)
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
			"instance_type": "redis-distributed-16",
		})
		data.SetId("udredis-1")

		if err := resourceUCloudRedisInstanceRead(data, newUmemTestRuntime(server.URL)); err != nil {
			t.Fatalf("read missing instance: %v", err)
		}
		if got := data.Id(); got != "" {
			t.Errorf("resource ID = %q, want empty for a missing instance", got)
		}
	})

	t.Run("API error keeps the ID and names it", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("DescribeURedisGroup", `{"RetCode":10001,"Message":"describe failed"}`)
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
			"instance_type": "redis-master-2",
		})
		data.SetId("uredis-1")

		err := resourceUCloudRedisInstanceRead(data, newUmemTestRuntime(server.URL))
		if err == nil {
			t.Fatal("expected the read to fail")
		}
		for _, want := range []string{"error on reading redis instance", "uredis-1", "describe failed"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("read error = %v, want it to contain %q", err, want)
			}
		}
		if got := data.Id(); got != "uredis-1" {
			t.Errorf("resource ID = %q, want the id to survive a read error", got)
		}
	})

	t.Run("empty id is a not found error", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		server := stub.start()
		defer server.Close()

		if _, err := newUmemTestClient(t, server.URL).describeActiveStandbyRedisById(""); !isNotFoundError(err) {
			t.Fatalf("describe with an empty id error = %v, want a not found error", err)
		}
	})
}

// TestDistributedRedisReadExpandsShardsProxiesAndReadMode covers the response
// expansion of the distributed read path, including the computed shard list, the
// proxy list, the read mode and the shard count written back to state.
func TestDistributedRedisReadExpandsShardsProxiesAndReadMode(t *testing.T) {
	stub := newUmemAPIStub(t)
	stub.on("DescribeUMemSpace", distributedRedisListBody(16, statusRunning))
	stub.on("DescribeUMemBlockInfo", blockInfoBody("master",
		block("udredis-1_0", 8, statusRunning), block("udredis-1_1", 8, statusRunning)))
	stub.on("DescribeUDRedisProxyInfo", proxyInfoBody(
		`{"ProxyId":"proxy-1","ResourceId":"udredis-1","State":"Running","Vip":"10.0.0.10"}`))
	server := stub.start()
	defer server.Close()

	data := newUmemTestData(t, resourceUCloudRedisInstance(), map[string]interface{}{
		"instance_type": "redis-distributed-16",
	})
	data.SetId("udredis-1")

	if err := resourceUCloudRedisInstanceRead(data, newUmemTestRuntime(server.URL)); err != nil {
		t.Fatalf("read distributed instance: %v", err)
	}

	// DescribeUMemBlockInfo and DescribeUDRedisProxyInfo need the instance zone.
	for _, action := range []string{"DescribeUMemBlockInfo", "DescribeUDRedisProxyInfo"} {
		if got := stub.calls(action)[0].Get("Zone"); got != "cn-sh2-01" {
			t.Errorf("%s Zone = %q, want cn-sh2-01", action, got)
		}
	}

	if got := data.Get("block_cnt").(int); got != 2 {
		t.Errorf("block_cnt = %d, want the live shard count 2", got)
	}
	if got := data.Get("read_mode"); got != "master" {
		t.Errorf("read_mode = %#v, want master", got)
	}
	if got := data.Get("instance_type"); got != "redis-distributed-16" {
		t.Errorf("instance_type = %#v, want redis-distributed-16", got)
	}

	blocks := data.Get("block_set").([]interface{})
	if len(blocks) != 2 {
		t.Fatalf("block_set length = %d, want 2", len(blocks))
	}
	first := blocks[0].(map[string]interface{})
	for field, want := range map[string]interface{}{
		"block_id":   "udredis-1_0",
		"block_size": 8,
		"block_port": 6379,
		"block_vip":  "10.0.0.9",
		"block_type": "master",
	} {
		if got := first[field]; got != want {
			t.Errorf("block_set.0.%s = %#v, want %#v", field, got, want)
		}
	}

	proxies := data.Get("proxy_set").([]interface{})
	if len(proxies) != 1 {
		t.Fatalf("proxy_set length = %d, want 1", len(proxies))
	}
	proxy := proxies[0].(map[string]interface{})
	for field, want := range map[string]interface{}{
		"proxy_id":    "proxy-1",
		"resource_id": "udredis-1",
		"state":       statusRunning,
		"vip":         "10.0.0.10",
	} {
		if got := proxy[field]; got != want {
			t.Errorf("proxy_set.0.%s = %#v, want %#v", field, got, want)
		}
	}
}

// TestMemcacheCRUDRequests covers the memcache resource at the same level:
// create request conversion, response expansion, the empty-id guard, NotFound on
// read and the delete call.
func TestMemcacheCRUDRequests(t *testing.T) {
	describeMemcache := func(size int, state string) string {
		return fmt.Sprintf(`{"RetCode":0,"TotalCount":1,"DataSet":[{
			"ResourceId":"umemcache-1","Name":"tf-acc-memcache","Zone":"cn-sh2-01","Tag":"Default",
			"VPCId":"uvnet-1","SubnetId":"subnet-1","Size":%d,"State":%q,"ChargeType":"Month",
			"CreateTime":1700000000,"ExpireTime":0,
			"Address":[{"IP":"10.0.0.11","Port":11211}]}]}`, size, state)
	}

	t.Run("create converts the request and expands the response", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("CreateUMemcacheGroup", `{"RetCode":0,"GroupId":"umemcache-1"}`)
		stub.on("DescribeUMem", describeMemcache(4, statusRunning))
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudMemcacheInstance(), map[string]interface{}{
			"availability_zone": "cn-sh2-01",
			"name":              "tf-acc-memcache",
			"instance_type":     "memcache-master-4",
			"charge_type":       "month",
			"duration":          1,
			"vpc_id":            "uvnet-1",
			"subnet_id":         "subnet-1",
		})

		if err := resourceUCloudMemcacheInstanceCreate(data, newUmemTestRuntime(server.URL)); err != nil {
			t.Fatalf("create memcache: %v", err)
		}

		create := stub.only(t, "CreateUMemcacheGroup")
		for field, want := range map[string]string{
			"Zone":     "cn-sh2-01",
			"VPCId":    "uvnet-1",
			"SubnetId": "subnet-1",
			"Size":     "4",
			"Protocol": "memcache",
			"Quantity": "1",
			"Name":     "tf-acc-memcache",
		} {
			if got := create.Get(field); got != want {
				t.Errorf("CreateUMemcacheGroup %s = %q, want %q", field, got, want)
			}
		}
		if got := data.Id(); got != "umemcache-1" {
			t.Errorf("resource ID = %q, want umemcache-1", got)
		}
		for field, want := range map[string]interface{}{
			"vpc_id":        "uvnet-1",
			"subnet_id":     "subnet-1",
			"status":        statusRunning,
			"instance_type": "memcache-master-4",
		} {
			if got := data.Get(field); got != want {
				t.Errorf("state %s = %#v, want %#v", field, got, want)
			}
		}
	})

	t.Run("create without a group id stores no id", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("CreateUMemcacheGroup", `{"RetCode":0}`)
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudMemcacheInstance(), map[string]interface{}{
			"availability_zone": "cn-sh2-01",
			"name":              "tf-acc-memcache",
			"instance_type":     "memcache-master-4",
			"charge_type":       "month",
			"vpc_id":            "uvnet-1",
			"subnet_id":         "subnet-1",
		})

		if err := resourceUCloudMemcacheInstanceCreate(data, newUmemTestRuntime(server.URL)); err == nil {
			t.Fatal("expected create to fail when the response carries no instance id")
		}
		if got := data.Id(); got != "" {
			t.Errorf("resource ID after failed create = %q, want empty", got)
		}
		if got := stub.count("DescribeUMem"); got != 0 {
			t.Errorf("created an empty id and polled the API %d times, want 0", got)
		}
	})

	t.Run("read of a missing instance clears the id", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("DescribeUMem", `{"RetCode":0,"TotalCount":0,"DataSet":[]}`)
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudMemcacheInstance(), map[string]interface{}{
			"instance_type": "memcache-master-4",
		})
		data.SetId("umemcache-1")

		if err := resourceUCloudMemcacheInstanceRead(data, newUmemTestRuntime(server.URL)); err != nil {
			t.Fatalf("read missing memcache: %v", err)
		}
		if got := data.Id(); got != "" {
			t.Errorf("resource ID = %q, want empty for a missing instance", got)
		}
	})

	t.Run("delete calls the API and stops once the instance is gone", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("DeleteUMemcacheGroup", `{"RetCode":0}`)
		stub.on("DescribeUMem", `{"RetCode":0,"TotalCount":0,"DataSet":[]}`)
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudMemcacheInstance(), map[string]interface{}{
			"instance_type": "memcache-master-4",
		})
		data.SetId("umemcache-1")

		if err := resourceUCloudMemcacheInstanceDelete(data, newUmemTestRuntime(server.URL)); err != nil {
			t.Fatalf("delete memcache: %v", err)
		}
		if got := stub.only(t, "DeleteUMemcacheGroup").Get("GroupId"); got != "umemcache-1" {
			t.Errorf("DeleteUMemcacheGroup GroupId = %q, want umemcache-1", got)
		}
	})

	t.Run("read error propagates with the resource id", func(t *testing.T) {
		stub := newUmemAPIStub(t)
		stub.on("DescribeUMem", `{"RetCode":10002,"Message":"memcache describe failed"}`)
		server := stub.start()
		defer server.Close()

		data := newUmemTestData(t, resourceUCloudMemcacheInstance(), map[string]interface{}{
			"instance_type": "memcache-master-4",
		})
		data.SetId("umemcache-1")

		err := resourceUCloudMemcacheInstanceRead(data, newUmemTestRuntime(server.URL))
		if err == nil {
			t.Fatal("expected the read to fail")
		}
		for _, want := range []string{"error on reading memcache instance", "umemcache-1", "memcache describe failed"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("read error = %v, want it to contain %q", err, want)
			}
		}
	})
}
