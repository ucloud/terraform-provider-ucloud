package ucloud

import (
	"encoding/json"
	"fmt"
	"github.com/ucloud/ucloud-sdk-go/external"
	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	pumem "github.com/ucloud/ucloud-sdk-go/private/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/ipsecvpn"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

// Config is the configuration of ucloud meta data
type Config struct {
	PublicKey             string
	PrivateKey            string
	Profile               string
	SharedCredentialsFile string

	Region     string
	ProjectId  string
	Insecure   bool
	BaseURL    string
	MaxRetries int
}

type CloudShellCredential struct {
	Active    bool   `json:"active"`
	Cookie    string `json:"cookie"`
	CSRFToken string `json:"csrf_token"`
}

// Client will returns a client with connections for all product
func (c *Config) Client() (*UCloudClient, error) {
	var err error
	var client UCloudClient
	var cfg *ucloud.Config
	var cred *auth.Credential

	client.region = c.Region
	client.projectId = c.ProjectId

	cfgV := ucloud.NewConfig()
	cfg = &cfgV

	// set common attributes (region, project id, etc ...)
	cfg.Region = c.Region
	cfg.ProjectId = c.ProjectId

	// enable auto retry with http/connection error
	cfg.MaxRetries = c.MaxRetries
	cfg.LogLevel = log.PanicLevel
	cfg.UserAgent = "Terraform-UCloud/1.19.0"
	cfg.BaseUrl = c.BaseURL

	if isAcc() {
		//set DebugLevel for acceptance test
		cfg.LogLevel = log.DebugLevel

		// excepted logging
		cfg.SetActionLevel("GetRegion", log.WarnLevel)
	}

	var cloudShellCredHandler ucloud.HttpRequestHandler
	if len(c.Profile) > 0 {
		// load public/private key from shared credential file
		cred, err = external.LoadUCloudCredentialFile(c.SharedCredentialsFile, c.Profile)
		if err != nil {
			return nil, fmt.Errorf("cannot load shared credential file, %s", err)
		}
	} else if len(c.PublicKey) > 0 && len(c.PrivateKey) > 0 {
		// load public/private key from shared credential file
		credV := auth.NewCredential()
		cred = &credV
		cred.PublicKey = c.PublicKey
		cred.PrivateKey = c.PrivateKey
	} else if v := os.Getenv("CLOUD_SHELL"); v == "true" {
		csCred := make([]CloudShellCredential, 0)
		// load credential from default cloud shell credential path
		if err := loadJSONFile(defaultCloudShellCredPath(), &csCred); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("cloud shell credential file is empty, %s", err)
			} else {
				return nil, fmt.Errorf("cannot load cloud shell credential file, %s", err)
			}
		}
		// get default cloud shell credential
		defaultCsCred := &CloudShellCredential{}
		for i := 0; i < len(csCred); i++ {
			if csCred[i].Active == true {
				defaultCsCred = &csCred[i]
				break
			}
		}
		// set cloud shell client handler
		cloudShellCredHandler = func(c *ucloud.Client, req *http.HttpRequest) (*http.HttpRequest, error) {
			req.SetHeader("Cookie", defaultCsCred.Cookie)
			req.SetHeader("Csrf-Token", defaultCsCred.CSRFToken)
			return req, nil
		}
	}

	// initialize client connections
	client.uhostconn = uhost.NewClient(cfg, cred)
	client.unetconn = unet.NewClient(cfg, cred)
	client.ulbconn = ulb.NewClient(cfg, cred)
	client.vpcconn = vpc.NewClient(cfg, cred)
	client.uaccountconn = uaccount.NewClient(cfg, cred)
	client.udiskconn = udisk.NewClient(cfg, cred)
	client.udpnconn = udpn.NewClient(cfg, cred)
	client.udbconn = udb.NewClient(cfg, cred)
	client.umemconn = umem.NewClient(cfg, cred)
	client.ipsecvpnClient = ipsecvpn.NewClient(cfg, cred)

	// initialize client connections for private usage
	client.pumemconn = pumem.NewClient(cfg, cred)

	if cloudShellCredHandler != nil {
		client.uhostconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.unetconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.ulbconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.vpcconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.uaccountconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.udiskconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.udpnconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.udbconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.umemconn.AddHttpRequestHandler(cloudShellCredHandler)
		client.ipsecvpnClient.AddHttpRequestHandler(cloudShellCredHandler)
		client.pumemconn.AddHttpRequestHandler(cloudShellCredHandler)
	}

	client.config = cfg
	client.credential = cred
	return &client, nil
}

func defaultCloudShellCredPath() string {
	return filepath.Join(userHomeDir(), ".ucloud", "credential.json")
}

func loadJSONFile(path string, p interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	c, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(c, p)
	if err != nil {
		return err
	}
	return nil
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
