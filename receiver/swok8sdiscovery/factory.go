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
	"context"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/internal/k8sconfig"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/receiver/swok8sdiscovery/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithLogs(
			createLogsReceiver,
			metadata.LogsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
	}
}

// Build-in configuration
func createImplicitConfig() component.Config {
	return &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
		Database: &DatabaseDiscoveryConfig{
			ImageRules: []*ImageRule{
				{
					DatabaseType: "redis",
					Patterns:     []string{".*/redis:.*"},
					DefaultPort:  6379,
				},
				{
					DatabaseType: "mysql",
					Patterns:     []string{".*/mysql:.*", ".*/mariadb:.*", ".*/mysql-server:.*"},
					DefaultPort:  3306,
				},
				{
					DatabaseType: "sqlserver",
					Patterns:     []string{".*/mssql:.*", ".*/mssql/server:.*"},
					DefaultPort:  1433,
				},
				{
					DatabaseType: "postgresql",
					Patterns:     []string{".*/postgres:.*", ".*/postgresql:.*"},
					DefaultPort:  5432,
				},
				{
					DatabaseType: "mongodb",
					Patterns:     []string{".*/mongo:.*"},
					DefaultPort:  27017,
				},
			},
			DomainRules: []*DomainRule{
				{
					DatabaseType: "redis",
					Patterns: []string{
						`\.redis\.cache\.windows\.net$`,
						`\.elasticache(?:\.[a-z]{2}(?:-[a-z]+){1,2}-\d+)?\.amazonaws\.com$`,
						`\.redis(?:\.[a-z]{2}(?:-[a-z]+){1,2}-\d+)?\.amazonaws\.com$`,
						`\.cache(?:\.[a-z]{2}(?:-[a-z]+){1,2}-\d+)?\.amazonaws\.com$`},
				},
				{
					DatabaseType: "mysql",
					Patterns: []string{
						`\.mysql\.database\.azure\.com$`,
						`\.rds(?:\.[a-z]{2}(?:-[a-z]+){1,2}-\d+)?\.amazonaws\.com$`},
					DomainHints: []string{"mysql", "mariadb"},
				},
				{
					DatabaseType: "sqlserver",
					Patterns: []string{
						`\.database\.windows\.net$`,
						`\.rds(?:\.[a-z]{2}(?:-[a-z]+){1,2}-\d+)?\.amazonaws\.com$`},
					DomainHints: []string{"sqlserver", "mssql"},
				},
				{
					DatabaseType: "postgresql",
					Patterns: []string{
						`\.postgres\.database\.azure\.com$`,
						`\.rds(?:\.[a-z]{2}(?:-[a-z]+){1,2}-\d+)?\.amazonaws\.com$`},
					DomainHints: []string{"postgres"},
				},
				{
					DatabaseType: "mongodb",
					Patterns: []string{
						`\.mongo\.cosmos\.azure\.com$`,
						`\.docdb(?:\.[a-z]{2}(?:-[a-z]+){1,2}-\d+)?\.amazonaws\.com$`},
				},
			},
		},
	}
}

func createLogsReceiver(
	_ context.Context,
	params receiver.Settings,
	cfg component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {
	rcfg := cfg.(*Config)
	return newReceiver(params, rcfg, consumer)
}
