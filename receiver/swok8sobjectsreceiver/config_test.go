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

package swok8sobjectsreceiver

import (
	"path/filepath"
	"testing"
	"time"

	"go.opentelemetry.io/collector/confmap/xconfmap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apiWatch "k8s.io/apimachinery/pkg/watch"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swok8sobjectsreceiver/internal/metadata"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id       component.ID
		expected *Config
	}{
		{
			id: component.NewIDWithName(metadata.Type, ""),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: k8sconfig.AuthTypeServiceAccount,
				},
				Objects: []*K8sObjectsConfig{
					{
						Name:          "pods",
						Mode:          PullMode,
						Interval:      time.Hour,
						FieldSelector: "status.phase=Running",
						LabelSelector: "environment in (production),tier in (frontend)",
						gvr: &schema.GroupVersionResource{
							Group:    "",
							Version:  "v1",
							Resource: "pods",
						},
					},
					{
						Name:            "events",
						Mode:            WatchMode,
						Namespaces:      []string{"default"},
						Group:           "events.k8s.io",
						ResourceVersion: "",
						ExcludeWatchType: []apiWatch.EventType{
							apiWatch.Deleted,
						},
						gvr: &schema.GroupVersionResource{
							Group:    "events.k8s.io",
							Version:  "v1",
							Resource: "events",
						},
					},
				},
				makeDiscoveryClient: getMockDiscoveryClient,
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "pull_with_resource"),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: k8sconfig.AuthTypeServiceAccount,
				},
				Objects: []*K8sObjectsConfig{
					{
						Name:            "pods",
						Mode:            PullMode,
						ResourceVersion: "1",
						Interval:        time.Hour,
						gvr: &schema.GroupVersionResource{
							Group:    "",
							Version:  "v1",
							Resource: "pods",
						},
					},
					{
						Name:     "events",
						Mode:     PullMode,
						Interval: time.Hour,
						gvr: &schema.GroupVersionResource{
							Group:    "",
							Version:  "v1",
							Resource: "events",
						},
					},
				},
				makeDiscoveryClient: getMockDiscoveryClient,
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "watch_with_resource"),
			expected: &Config{
				APIConfig: k8sconfig.APIConfig{
					AuthType: k8sconfig.AuthTypeServiceAccount,
				},
				Objects: []*K8sObjectsConfig{
					{
						Name:            "events",
						Mode:            WatchMode,
						Namespaces:      []string{"default"},
						Group:           "events.k8s.io",
						ResourceVersion: "",
						gvr: &schema.GroupVersionResource{
							Group:    "events.k8s.io",
							Version:  "v1",
							Resource: "events",
						},
					},
					{
						Name:            "events",
						Mode:            WatchMode,
						Namespaces:      []string{"default"},
						Group:           "events.k8s.io",
						ResourceVersion: "2",
						gvr: &schema.GroupVersionResource{
							Group:    "events.k8s.io",
							Version:  "v1",
							Resource: "events",
						},
					},
				},
				makeDiscoveryClient: getMockDiscoveryClient,
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "invalid_resource"),
		},
		{
			id: component.NewIDWithName(metadata.Type, "exclude_deleted_with_pull"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
			require.NoError(t, err)

			factory := NewFactory()
			cfg := factory.CreateDefaultConfig().(*Config)
			cfg.makeDiscoveryClient = getMockDiscoveryClient

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			if tt.expected == nil {
				err = xconfmap.Validate(cfg)
				assert.Error(t, err)
				return
			}
			assert.NoError(t, xconfmap.Validate(cfg))
			assert.Equal(t, tt.expected.AuthType, cfg.AuthType)
			assert.Equal(t, tt.expected.Objects, cfg.Objects)
		})
	}
}

func TestValidateResourceConflict(t *testing.T) {
	mockClient := newMockDynamicClient()
	rCfg := createDefaultConfig().(*Config)
	rCfg.makeDynamicClient = mockClient.getMockDynamicClient
	rCfg.makeDiscoveryClient = getMockDiscoveryClient

	// Validate it should choose first gvr if group is not specified
	rCfg.Objects = []*K8sObjectsConfig{
		{
			Name: "myresources",
			Mode: PullMode,
		},
	}

	err := rCfg.Validate()
	require.NoError(t, err)
	assert.Equal(t, "group1", rCfg.Objects[0].gvr.Group)

	// Validate it should choose gvr for specified group
	rCfg.Objects = []*K8sObjectsConfig{
		{
			Name:  "myresources",
			Mode:  PullMode,
			Group: "group2",
		},
	}

	err = rCfg.Validate()
	require.NoError(t, err)
	assert.Equal(t, "group2", rCfg.Objects[0].gvr.Group)
}

func TestClientRequired(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)
	err := rCfg.Validate()
	require.Error(t, err)
}
