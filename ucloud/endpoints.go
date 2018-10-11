package ucloud

type endpoint string

const (
	publicEndpoint         endpoint = "http://api.ucloud.cn"
	publicInsecureEndpoint endpoint = "https://api.ucloud.cn"
)

// GetURL will return endpoint as string
func (e endpoint) GetURL() string {
	return string(e)
}

// GetEndpointURL will return endpoint url string by region
func GetEndpointURL(region string) string {
	return publicEndpoint.GetURL()
}

// GetInsecureEndpointURL will return endpoint url string by region
func GetInsecureEndpointURL(region string) string {
	return publicInsecureEndpoint.GetURL()
}
