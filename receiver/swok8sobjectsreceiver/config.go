// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Source: https://github.com/open-telemetry/opentelemetry-collector-contrib
// Changes customizing the original source code

package swok8sobjectsreceiver // import "github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swok8sobjectsreceiver"

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/collector/component"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apiWatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
)

type mode string

const (
	PullMode  mode = "pull"
	WatchMode mode = "watch"

	defaultPullInterval    time.Duration = time.Hour
	defaultMode            mode          = PullMode
	defaultResourceVersion               = "1"
)

var modeMap = map[mode]bool{
	PullMode:  true,
	WatchMode: true,
}

type K8sObjectsConfig struct {
	Name             string               `mapstructure:"name"`
	Group            string               `mapstructure:"group"`
	Namespaces       []string             `mapstructure:"namespaces"`
	Mode             mode                 `mapstructure:"mode"`
	LabelSelector    string               `mapstructure:"label_selector"`
	FieldSelector    string               `mapstructure:"field_selector"`
	Interval         time.Duration        `mapstructure:"interval"`
	ResourceVersion  string               `mapstructure:"resource_version"`
	ExcludeWatchType []apiWatch.EventType `mapstructure:"exclude_watch_type"`
	exclude          map[apiWatch.EventType]bool
	gvr              *schema.GroupVersionResource
}

type Config struct {
	k8sconfig.APIConfig `mapstructure:",squash"`

	Objects   []*K8sObjectsConfig `mapstructure:"objects"`
	StorageID *component.ID       `mapstructure:"storage"`

	// For mocking purposes only.
	makeDiscoveryClient func() (discovery.ServerResourcesInterface, error)
	makeDynamicClient   func() (dynamic.Interface, error)
}

func (c *Config) Validate() error {
	validObjects, err := c.getValidObjects()
	if err != nil {
		return err
	}
	for _, object := range c.Objects {
		gvrs, ok := validObjects[object.Name]
		if !ok {
			availableResource := make([]string, len(validObjects))
			for k := range validObjects {
				availableResource = append(availableResource, k)
			}
			return fmt.Errorf("resource %v not found. Valid resources are: %v", object.Name, availableResource)
		}

		gvr := gvrs[0]
		for i := range gvrs {
			if gvrs[i].Group == object.Group {
				gvr = gvrs[i]
				break
			}
		}

		if object.Mode == "" {
			object.Mode = defaultMode
		} else if _, ok := modeMap[object.Mode]; !ok {
			return fmt.Errorf("invalid mode: %v", object.Mode)
		}

		if object.Mode == PullMode && object.Interval == 0 {
			object.Interval = defaultPullInterval
		}

		if object.Mode == PullMode && len(object.ExcludeWatchType) != 0 {
			return errors.New("the Exclude config can only be used with watch mode")
		}

		object.gvr = gvr
	}
	return nil
}

func (c *Config) getDiscoveryClient() (discovery.ServerResourcesInterface, error) {
	if c.makeDiscoveryClient != nil {
		return c.makeDiscoveryClient()
	}

	client, err := k8sconfig.MakeClient(c.APIConfig)
	if err != nil {
		return nil, err
	}

	return client.Discovery(), nil
}

func (c *Config) getDynamicClient() (dynamic.Interface, error) {
	if c.makeDynamicClient != nil {
		return c.makeDynamicClient()
	}

	return k8sconfig.MakeDynamicClient(c.APIConfig)
}

func (c *Config) getValidObjects() (map[string][]*schema.GroupVersionResource, error) {
	dc, err := c.getDiscoveryClient()
	if err != nil {
		return nil, err
	}

	res, err := dc.ServerPreferredResources()
	if err != nil {
		// Check if Partial result is returned from discovery client, that means some API servers have issues,
		// but we can still continue, as we check for the needed groups later in Validate function.
		if res != nil && !discovery.IsGroupDiscoveryFailedError(err) {
			return nil, err
		}
	}

	validObjects := make(map[string][]*schema.GroupVersionResource)

	for _, group := range res {
		split := strings.Split(group.GroupVersion, "/")
		if len(split) == 1 && group.GroupVersion == "v1" {
			split = []string{"", "v1"}
		}
		for _, resource := range group.APIResources {
			validObjects[resource.Name] = append(validObjects[resource.Name], &schema.GroupVersionResource{
				Group:    split[0],
				Version:  split[1],
				Resource: resource.Name,
			})
		}
	}
	return validObjects, nil
}
