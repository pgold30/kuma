package mesh

import (
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (d *HealthCheckResource) Validate() error {
	var err validators.ValidationError
	err.Add(d.validateSources())
	err.Add(d.validateDestinations())
	err.Add(d.validateConf())
	return err.OrNil()
}

func (d *HealthCheckResource) validateSources() validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("sources"), d.Spec.Sources, ValidateSelectorsOpts{
		RequireAtLeastOneSelector: true,
		ValidateSelectorOpts: ValidateSelectorOpts{
			RequireAtLeastOneTag: true,
			RequireService:       true,
		},
	})
}

func (d *HealthCheckResource) validateDestinations() (err validators.ValidationError) {
	return ValidateSelectors(validators.RootedAt("destinations"), d.Spec.Destinations, OnlyServiceTagAllowed)
}

func (d *HealthCheckResource) validateConfHttpPath(
	path validators.PathBuilder,
) (err validators.ValidationError) {
	httpConf := d.Spec.Conf.GetHttp()

	if httpConf.Path == "" {
		err.AddViolationAt(path, "has to be defined and cannot be empty")
	}

	return
}

func (d *HealthCheckResource) validateConfHttpRequestHeadersToAdd(
	path validators.PathBuilder,
) (err validators.ValidationError) {
	httpConf := d.Spec.Conf.GetHttp()

	for i, header := range httpConf.RequestHeadersToAdd {
		path := path.Index(i).Field("header")

		if header.Header == nil {
			err.AddViolationAt(path, "has to be defined")
			continue
		}

		if header.Header.Key == "" {
			err.AddViolationAt(path.Field("key"), "cannot be empty")
		}
	}

	return
}

func (d *HealthCheckResource) validateConfHttpExpectedStatuses(
	path validators.PathBuilder,
) (err validators.ValidationError) {
	httpConf := d.Spec.Conf.GetHttp()

	if httpConf.ExpectedStatuses != nil {
		for i, status := range httpConf.ExpectedStatuses {
			if status.Value < 100 || status.Value >= 600 {
				err.AddViolationAt(
					path.Index(i),
					"must be in range [100, 600)",
				)
			}
		}
	}

	return
}

func (d *HealthCheckResource) validateConfHttp(
	path validators.PathBuilder,
) (err validators.ValidationError) {
	err.Add(d.validateConfHttpPath(path.Field("path")))
	err.Add(d.validateConfHttpExpectedStatuses(path.Field("expectedStatuses")))
	err.Add(d.validateConfHttpRequestHeadersToAdd(path.Field("requestHeadersToAdd")))
	return
}

func (d *HealthCheckResource) validateConf() (err validators.ValidationError) {
	path := validators.RootedAt("conf")
	if d.Spec.GetConf() == nil {
		err.AddViolationAt(path, "has to be defined")
		return
	}
	err.Add(ValidateDuration(path.Field("interval"), d.Spec.Conf.Interval))
	err.Add(ValidateDuration(path.Field("timeout"), d.Spec.Conf.Timeout))
	err.Add(ValidateThreshold(path.Field("unhealthyThreshold"), d.Spec.Conf.UnhealthyThreshold))
	err.Add(ValidateThreshold(path.Field("healthyThreshold"), d.Spec.Conf.HealthyThreshold))
	if d.Spec.Conf.InitialJitter != nil {
		err.Add(ValidateDuration(path.Field("initialJitter"), d.Spec.Conf.InitialJitter))
	}
	if d.Spec.Conf.IntervalJitter != nil {
		err.Add(ValidateDuration(path.Field("intervalJitter"), d.Spec.Conf.IntervalJitter))
	}
	if d.Spec.Conf.GetHttp() != nil {
		err.Add(d.validateConfHttp(path.Field("http")))
	}
	return
}
