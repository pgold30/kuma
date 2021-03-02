package xds

import (
	"context"
	"github.com/go-logr/logr"
	v1 "github.com/kumahq/kuma/app/kuma-prometheus-sd/pkg/discovery/xds/v1"
	"github.com/kumahq/kuma/app/kuma-prometheus-sd/pkg/discovery/xds/v1alpha1"
	"github.com/pkg/errors"

	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
)

// DiscovererE is like discovery.Discoverer but allows returning errors
type DiscovererE interface {
	Run(ctx context.Context, ch chan<- []*targetgroup.Group) error
}

// DiscovererFactory
type DiscovererFactory interface {
	CreateDiscoverer(config DiscoveryConfig, log logr.Logger) DiscovererE
}

type discoverer struct {
	factory DiscovererFactory
	log     logr.Logger
	config  DiscoveryConfig
}

func NewDiscoverer(config DiscoveryConfig, log logr.Logger) (discovery.Discoverer, error) {
	var factory DiscovererFactory
	switch config.ApiVersion {
	case V1:
		factory = v1.NewFactory()
	case V1_ALPHA1:
		factory = v1alpha1.NewFactory()
	default:
		return nil, errors.Errorf("invalid MADS apiVersion %s", config.ApiVersion)
	}

	return &discoverer{
		factory: factory,
		log:     log,
		config:  config,
	}, nil
}

// Run implements discovery.Discoverer interface.
func (d *discoverer) Run(ctx context.Context, ch chan<- []*targetgroup.Group) {
	// notice that Prometheus discovery.Discoverer abstraction doesn't allow failures,
	// so we must ensure that xDS client is up-and-running all the time.
	for streamID := uint64(1); ; streamID++ {
		logger := d.log.WithValues("streamID", streamID)
		errCh := make(chan error, 1)
		go func(errCh chan<- error) {
			defer close(errCh)
			// recover from a panic
			defer func() {
				if e := recover(); e != nil {
					if err, ok := e.(error); ok {
						errCh <- err
					} else {
						errCh <- errors.Errorf("%v", e)
					}
				}
			}()
			stream := d.factory.CreateDiscoverer(d.config, logger)
			errCh <- stream.Run(ctx, ch)
		}(errCh)
		select {
		case <-ctx.Done():
			logger.Info("done")
			break
		case err := <-errCh:
			if err != nil {
				logger.Error(err, "xDS stream terminated with an error")
			}
		}
	}
}
