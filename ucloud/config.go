package ucloud

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
)

// Config is the configuration of ucloud meta data
type Config struct {
	PublicKey  string
	PrivateKey string
	Region     string
	ProjectId  string

	MaxRetries int
	Insecure   bool
}

// UCloudClient is the ucloud openapi client
type UCloudClient struct {
	region    string
	projectId string

	uhostconn    *uhost.UHostClient
	unetconn     *unet.UNetClient
	ulbconn      *ulb.ULBClient
	vpcconn      *vpc.VPCClient
	uaccountconn *uaccount.UAccountClient
	udiskconn    *udisk.UDiskClient
	udpnconn     *udpn.UDPNClient
}

// Client will returns a client with connections for all product
func (c *Config) Client() (*UCloudClient, error) {
	var client UCloudClient
	client.region = c.Region
	client.projectId = c.ProjectId

	// set common attributes (region, project id, etc ...)
	config := ucloud.NewConfig()
	config.Region = c.Region
	config.ProjectId = c.ProjectId

	// enable auto retry with http/connection error
	config.MaxRetries = c.MaxRetries
	config.LogLevel = log.DebugLevel
	config.UserAgent = "Terraform/1.11.x"

	// set endpoint with insecure https connection
	if c.Insecure {
		config.BaseUrl = GetInsecureEndpointURL(c.Region)
	} else {
		config.BaseUrl = GetEndpointURL(c.Region)
	}

	// credential with publicKey/privateKey
	credential := auth.NewCredential()
	credential.PublicKey = c.PublicKey
	credential.PrivateKey = c.PrivateKey

	// initialize client connections
	client.uhostconn = uhost.NewClient(&config, &credential)
	client.unetconn = unet.NewClient(&config, &credential)
	client.ulbconn = ulb.NewClient(&config, &credential)
	client.vpcconn = vpc.NewClient(&config, &credential)
	client.uaccountconn = uaccount.NewClient(&config, &credential)
	client.udiskconn = udisk.NewClient(&config, &credential)
	client.udpnconn = udpn.NewClient(&config, &credential)

	return &client, nil
}
