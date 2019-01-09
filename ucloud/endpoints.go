package ucloud

type endpoint string

const (
	publicInsecureEndpoint endpoint = "http://api.ucloud.cn"
	publicSecureEndpoint   endpoint = "https://api.ucloud.cn"
)

// GetURL will return endpoint as string
func (e endpoint) GetURL() string {
	return string(e)
}

// GetInsecureEndpointURL will return endpoint url string by region
func GetInsecureEndpointURL(region string) string {
	return publicInsecureEndpoint.GetURL()
}

// GetEndpointURL will return endpoint url string by region
func GetEndpointURL(region string) string {
	return publicSecureEndpoint.GetURL()
}
