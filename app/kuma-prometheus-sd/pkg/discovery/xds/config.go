package xds


type ApiVersion string

const (
	V1_ALPHA1 = ApiVersion("v1alpha1")
	V1 = ApiVersion("v1")
)

type DiscoveryConfig struct {
	ServerURL  string
	ClientName string
	ApiVersion ApiVersion
}
