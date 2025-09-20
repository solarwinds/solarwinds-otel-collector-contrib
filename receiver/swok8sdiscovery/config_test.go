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

package swok8sdiscovery

import (
	"path/filepath"
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swok8sdiscovery/internal/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/confmap/xconfmap"
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
				Interval: defaultInterval,
				Database: &DatabaseDiscoveryConfig{
					// add data from testdata/config.yaml
					ImageRules: []ImageRule{
						{
							DatabaseType: "mysql",
							Patterns:     []string{"mysql*", "mariadb*"},
							DefaultPort:  3306,
						},
						{
							DatabaseType: "postgres",
							Patterns:     []string{"postgres*", "postgresql*"},
							DefaultPort:  5432,
						},
					},
					DomainRules: []DomainRule{
						{
							DatabaseType: "mysql",
							Patterns:     []string{"mysql*", "mariadb*"},
						},
						{
							DatabaseType: "postgres",
							Patterns: []string{
								`\.postgres\.database\.azure\.com$`,
								`\.rds(?:\.[a-z]{2}(?:-[a-z]+){1,2}-\d+)?\.amazonaws\.com$`,
							},
							DomainHints: []string{"postgres"},
						},
					},
				},
			},
		},
		{
			id: component.NewIDWithName(metadata.Type, "missing_domain_rules"),
		},
		{
			id: component.NewIDWithName(metadata.Type, "missing_image_rules"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
			require.NoError(t, err)

			factory := NewFactory()
			cfg := factory.CreateDefaultConfig().(*Config)

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
			assert.Equal(t, tt.expected.Interval, cfg.Interval)

			require.Equal(t, len(tt.expected.Database.ImageRules), len(cfg.Database.ImageRules))
			for i := range tt.expected.Database.ImageRules {
				assert.Equal(t, tt.expected.Database.ImageRules[i].DatabaseType, cfg.Database.ImageRules[i].DatabaseType)
				assert.Equal(t, tt.expected.Database.ImageRules[i].Patterns, cfg.Database.ImageRules[i].Patterns)
				assert.Equal(t, tt.expected.Database.ImageRules[i].DefaultPort, cfg.Database.ImageRules[i].DefaultPort)
				assert.Len(t, cfg.Database.ImageRules[i].PatternsCompiled, len(cfg.Database.ImageRules[i].Patterns), "image_rules[%d].MatchesCompiled length mismatch", i)
			}

			require.Equal(t, len(tt.expected.Database.DomainRules), len(cfg.Database.DomainRules))
			for i := range tt.expected.Database.DomainRules {
				assert.Equal(t, tt.expected.Database.DomainRules[i].DatabaseType, cfg.Database.DomainRules[i].DatabaseType)
				assert.Equal(t, tt.expected.Database.DomainRules[i].Patterns, cfg.Database.DomainRules[i].Patterns)
				assert.Len(t, cfg.Database.DomainRules[i].PatternsCompiled, len(cfg.Database.DomainRules[i].Patterns), "domain_rules[%d].MatchesCompiled length mismatch", i)
			}
		})
	}
}

func TestDefaultConfigFails(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)
	err := rCfg.Validate()
	require.NoError(t, err)
}
