package v1

import (
	"fmt"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/util/strutil"

	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
)

const (
	// meshLabel is the name of the label that holds the mesh name.
	meshLabel = model.LabelName("mesh")
	// dataplaneLabel is the name of the label that holds the dataplane name.
	dataplaneLabel = model.LabelName("dataplane")
)

type Converter struct{}

func (c Converter) ConvertAll(assignments []*observability_v1.MonitoringAssignment) []*targetgroup.Group {
	var groups []*targetgroup.Group
	for _, assignment := range assignments {
		groups = append(groups, c.Convert(assignment)...)
	}
	return groups
}

// Beware of the following constraints when it comes to integration with Prometheus:
//
//  1. Prometheus model for all `sd`s except for `file_sd` looks like this:
//
//     // Group is a set of targets with a common label set(production , test, staging etc.).
//     type Group struct {
//         // Targets is a list of targets identified by a label set. Each target is
//         // uniquely identifiable in the group by its address label.
//         Targets []model.LabelSet
//         // Labels is a set of labels that is common across all targets in the group.
//         Labels model.LabelSet
//
//         // Source is an identifier that describes a group of targets.
//         Source string
//     }
//
//     That is why Kuma's MonitoringAssignment was designed to be close to that model.
//
//  2. However, `file_sd` uses different model for reading data from a file:
//
//     struct {
//         Targets []string       `yaml:"targets"`
//         Labels  model.LabelSet `yaml:"labels"`
//     }
//
//     Notice that Targets is just a list of addresses rather than a list of model.LabelSet.
//
//  3. Because of that mismatch, some form of conversion is unavoidable on client side,
//     e.g. inside `kuma-prometheus-sd`
//
//  4. The next component that imposes its constraints is `custom-sd`- adapter
//     (https://github.com/prometheus/prometheus/tree/master/documentation/examples/custom-sd)
//     that is recommended for use by all `file_sd`-based `sd`s.
//
//     This adapter is doing conversion from Prometheus model into `file_sd` model
//     and it expects that `Targets` field has only 1 label - `__address__` -
//     and the rest of the labels must be a part of `Labels` field.
//
//  5. Therefore, we need to convert MonitoringAssignment into a model that `custom-sd` expects.
//
//  In practice, it means that generated MonitoringAssignment will look the following way:
//
//     name: /meshes/default/dataplanes/backend-01
//     targets:
//     - labels:
//         __address__: 192.168.0.1:8080
//     labels:
//       __scheme__: http
//       __metrics_path__: /metrics
//       job: backend
//       instance: backend-01
//       mesh: default
//       dataplane: backend-01
//       env: prod
//       envs: ,prod,
//       service: backend
//       services: ,backend,
//
// assignment => []Group
// target => Group
func (c Converter) Convert(assignment *observability_v1.MonitoringAssignment) []*targetgroup.Group {
	var groups []*targetgroup.Group
	commonLabels := c.convertLabels(assignment.Labels)

	for i, target := range assignment.Targets {
		targetLabels := c.convertLabels(target.Labels)
		allLabels := commonLabels.Clone().Merge(targetLabels)

		allLabels[meshLabel] = model.LabelValue(assignment.Mesh)
		allLabels[dataplaneLabel] = model.LabelValue(assignment.Name)
		allLabels[model.JobLabel] = model.LabelValue(assignment.Service)
		allLabels[model.InstanceLabel] = model.LabelValue(assignment.Name)
		// workaround for custom_sd Adapter, only Target label is address
		allLabels[model.SchemeLabel] = model.LabelValue(target.Scheme)
		allLabels[model.MetricsPathLabel] = model.LabelValue(target.MetricsPath)

		group := &targetgroup.Group{
			Source: c.subSourceName(assignment.Name, i),
			Targets: []model.LabelSet{{
				model.AddressLabel: model.LabelValue(target.Address),
			}},
			Labels: allLabels,
		}
		groups = append(groups, group)
	}
	return groups
}

func (c Converter) convertLabels(labels map[string]string) model.LabelSet {
	labelSet := model.LabelSet{}
	for key, value := range labels {
		name := strutil.SanitizeLabelName(key)
		labelSet[model.LabelName(name)] = model.LabelValue(value)
	}
	return labelSet
}

func (c Converter) subSourceName(source string, i int) string {
	return fmt.Sprintf("%s/%d", source, i)
}
