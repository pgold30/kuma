package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
)

// VirtualHostConfigurer is responsible for configuring a single aspect of the entire Envoy VirtualHost,
// such as Route, CORS, etc.
type VirtualHostConfigurer interface {
	// Configure configures a single aspect on a given Envoy VirtualHost.
	Configure(virtualHost *envoy_route.VirtualHost) error
}

// RouteConfigurationConfigurer is responsible for configuring a single aspect of the entire Envoy RouteConfiguration,
// such as VirtualHost, HTTP headers to add or remove, etc.
type RouteConfigurationConfigurer interface {
	// Configure configures a single aspect on a given Envoy RouteConfiguration.
	Configure(routeConfiguration *envoy_api.RouteConfiguration) error
}
