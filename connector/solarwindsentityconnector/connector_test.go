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

package solarwindsentityconnector

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/plogtest"
	"github.com/solarwinds/solarwinds-otel-collector-contrib/connector/solarwindsentityconnector/internal/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/connector/connectortest"
	"go.opentelemetry.io/collector/consumer/consumertest"
)

func TestLogsToLogs(t *testing.T) {

	testCases := []struct {
		name     string
		folder   string
		noOutput bool
	}{
		{
			// ~~~ DESCRIPTION ~~~
			// Testing that more advanced conditions get evaluated correctly.
			// Checks that events for entities and relationships are still sent.
			name:   "when same type relationship has valid advanced condition, log event is sent",
			folder: "relationship/same-type-relationship/02-advanced-conditions",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// Checks that same type relationship for AWS EC2 is sent, together with the two AWS EC2 entities.
			name:   "when relationship for same type is inferred log event is sent",
			folder: "relationship/same-type-relationship/01-happy-path",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// Checks that if one of the two entities is not created, the relationship is not sent.
			name:   "when relationship for same type is not inferred no log is sent",
			folder: "relationship/same-type-relationship/04-no-match",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// Checks that same type relationship for AWS EC2 is sent, together with entities.
			// Input contains 3 resource logs
			// Uses prefixes, same as all same type relationship tests
			name:   "when entity is inferred log event is sent",
			folder: "entity/01-happy-path",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// Input is sending insufficient attributes for entity creation
			// Nothing should be sent
			name:   "when entity is not inferred no log is sent",
			folder: "entity/03-no-match",
		},
		{
			// ~~~ DESCRIPTION ~~~
			name:   "when relationship for same type and having common id attributes is inferred log event is sent",
			folder: "relationship/same-type-relationship/03-common-atr",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// config.yaml
			//		setup for Snowflake AWS EC2 and Pod entities, and relationship between Snowflake and AWS EC2
			// input.yaml
			//		contains extra k8s attribute not necessary for relationship, but not enough for Pod entity creation
			// expected-output.yaml
			//		log event is sent for Snowflake and EC2 entities, and relationship between them. No Pod entity.
			name:   "when log for different type relationship has redundant attributes, log event is sent",
			folder: "relationship/different-types-relationship/05-redundant-atr",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// config.yaml
			// 		is set up for Snowflake and AWS EC2 entities, with relationship between them
			// input.yaml
			//  	is missing "id" attribute for entity "AWS EC2"
			// expected-output.yaml
			// 		log event is sent but only for entity "Snowflake", no relationship log event is sent
			name:   "when log for different type relationship hasn't all necessary id attributes, log event is sent",
			folder: "relationship/different-types-relationship/03-missing-atr",
		},
		{
			// ~~~ DESCRIPTION ~~~
			name:   "when log for same type relationship, log event is sent with relationship attributes",
			folder: "relationship/same-type-relationship/05-res-atr",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// config.yaml
			// 		is set up for Snowflake and Pod entities, with relationship between them
			// 		relationship has set attributes (that should be sent with the relationship log event)
			// input.yaml
			//  	setup for entities and relationship to be sent, with the extra attribute for relationship
			// expected-output.yaml
			// 		relationship should be sent with the extra attribute, and also 2 entity log events
			name:   "when log for different type relationship, log event is sent with relationship attributes",
			folder: "relationship/different-types-relationship/07-res-atr",
		},
		{
			// ~~~ DESCRIPTION ~~~
			name:   "when log for entity has no valid condition, no log event is sent",
			folder: "entity/04-no-valid-condition",
		},
		{
			// ~~~ DESCRIPTION ~~~
			name:   "when log for entity has valid condition, log event is sent",
			folder: "entity/02-valid-condition",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// config.yaml
			// 		is set up for Pod and Cluster entities, with relationship between them
			// input.yaml
			//		does not fulfill the condition for relationship, and condition for Cluster entity
			// expected-output.yaml
			//		only entity log events for Pod is sent
			name:   "when log for different type relationship has no valid condition, no log relationship event is sent",
			folder: "relationship/different-types-relationship/04-no-valid-condition",
		},
		{
			// ~~~ DESCRIPTION ~~~
			name:   "when log for different type relationship has valid condition, log relationship event is sent",
			folder: "relationship/different-types-relationship/06-valid-condition",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// config.yaml
			// 		contains source_prefix:"src." & destination_prefix:"dst."
			// input.yaml
			//  	attributes are prefixed
			//      there are no unprefixed attribute copies, so no entity updates should happen
			// expected-output.yaml
			// 		only relationship update is sent (1 log record)
			name:   "different type relationship works with prefixes",
			folder: "relationship/different-types-relationship/01-with-prefixes",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// config.yaml
			// 		does not contain source_prefix & destination_prefix
			// 		due to this, there should be match on entity attributes and relationship attributes (3 log records)
			// input.yaml
			//  	attributes are not prefixed
			// expected-output.yaml
			// 		two entity updates are sent, and one relationship update is sent (3 log records)
			name:   "different type relationship works without prefixes",
			folder: "relationship/different-types-relationship/02-without-prefixes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := loadConfigFromFile(t, filepath.Join("testdata", "logsToLogs", tc.folder, "config.yaml"))
			require.NoError(t, err)
			factory := NewFactory()
			sink := &consumertest.LogsSink{}
			conn, err := factory.CreateLogsToLogs(context.Background(),
				connectortest.NewNopSettings(metadata.Type), cfg, sink)
			require.NoError(t, err)
			require.NotNil(t, conn)

			require.NoError(t, conn.Start(context.Background(), componenttest.NewNopHost()))
			defer func() {
				assert.NoError(t, conn.Shutdown(context.Background()))
			}()

			inputFile := filepath.Join("testdata", "logsToLogs", tc.folder, "input.yaml")
			testLogs, err := golden.ReadLogs(inputFile)

			assert.NoError(t, err)
			assert.NoError(t, conn.ConsumeLogs(context.Background(), testLogs))

			allLogs := sink.AllLogs()
			expectedFile := filepath.Join("testdata", "logsToLogs", tc.folder, "expected-output.yaml")

			if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
				assert.Len(t, allLogs, 0)
				return
			}

			expected, err := golden.ReadLogs(expectedFile)

			assert.NoError(t, err)
			assert.Equal(t, allLogs[0].LogRecordCount(), expected.LogRecordCount())
			assert.NoError(t, plogtest.CompareLogs(expected, allLogs[0], plogtest.IgnoreObservedTimestamp()))
		})
	}
}

func TestMetricsToLogs(t *testing.T) {
	testCases := []struct {
		name         string
		inputFile    string
		expectedFile string
		configFile   string
	}{
		{
			name:         "when same type relationship has valid advanced condition, log event is sent",
			inputFile:    "relationship/same-type-relationship/input-metric-same-type-relationship-advanced-conditions.yaml",
			expectedFile: "relationship/same-type-relationship/expected-metric-same-type-relationship-advanced-conditions.yaml",
		},
		{
			name:         "when entity is inferred, log event is sent",
			inputFile:    "entity/input-metric.yaml",
			expectedFile: "entity/expected-metric.yaml",
		},
		{
			name:      "when entity is not inferred, no log is sent",
			inputFile: "entity/input-metric-nomatch.yaml",
		},
		{
			name:         "when relationship for same type is inferred log event is sent",
			inputFile:    "relationship/same-type-relationship/input-metric-same-type-relationship.yaml",
			expectedFile: "relationship/same-type-relationship/expected-metric-same-type-relationship.yaml",
		},
		{
			name:         "when relationship for same type is not inferred no log is sent",
			inputFile:    "relationship/same-type-relationship/input-metric-same-type-relationship-nomatch.yaml",
			expectedFile: "relationship/same-type-relationship/expected-metric-same-type-relationship-nomatch.yaml",
		},

		{
			name:         "when relationship for same type and having common id attributes is inferred log event is sent",
			inputFile:    "relationship/same-type-relationship/input-metric-same-type-relationship-common-atr.yaml",
			expectedFile: "relationship/same-type-relationship/expected-metric-same-type-relationship-common-atr.yaml",
		},
		{
			name:         "when metric for different type relationship has redundant attributes, log event is sent",
			inputFile:    "relationship/different-types-relationship/input-metric-different-type-relationship-redundant-atr.yaml",
			expectedFile: "relationship/different-types-relationship/expected-metric-different-type-relationship-redundant-atr.yaml",
		},
		{
			name:         "when metric for different type relationship hasn't all necessary id attributes, log event is sent",
			inputFile:    "relationship/different-types-relationship/input-metric-different-type-relationship-missing-atr.yaml",
			expectedFile: "relationship/different-types-relationship/expected-metric-different-type-relationship-missing-atr.yaml",
		},
		{
			name:         "when metric for same type relationship, log event is sent with relationship attributes",
			inputFile:    "relationship/same-type-relationship/input-metric-same-type-relationship-res-atr.yaml",
			expectedFile: "relationship/same-type-relationship/expected-metric-same-type-relationship-res-atr.yaml",
		},
		{
			name:         "when metric for different type relationship, log event is sent with relationship attributes",
			inputFile:    "relationship/different-types-relationship/input-metric-different-type-relationship-res-atr.yaml",
			expectedFile: "relationship/different-types-relationship/expected-metric-different-type-relationship-res-atr.yaml",
		},
		{
			name:      "when metric for entity has no valid condition, no log event is sent",
			inputFile: "entity/input-metric-no-valid-condition.yaml",
		},
		{
			name:         "when metric for entity has valid condition, log event is sent",
			inputFile:    "entity/input-metric-valid-condition.yaml",
			expectedFile: "entity/expected-metric-valid-condition.yaml",
		},
		{
			name:         "when metric for different type relationship has no valid condition, no log relationship event is sent",
			inputFile:    "relationship/different-types-relationship/input-metric-different-type-relationship-no-valid-condition.yaml",
			expectedFile: "relationship/different-types-relationship/expected-metric-different-type-relationship-no-valid-condition.yaml",
		},
		{
			name:         "when metric for different type relationship has valid condition, log relationship event is sent",
			inputFile:    "relationship/different-types-relationship/input-metric-different-type-relationship-valid-condition.yaml",
			expectedFile: "relationship/different-types-relationship/expected-metric-different-type-relationship-valid-condition.yaml",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// config.yaml
			// 		contains source_prefix:"src." & destination_prefix:"dst."
			// input.yaml
			//  	attributes are prefixed
			//      there are no unexpected attribute copies, so no entity updates should happen
			// expected-output.yaml
			// 		only relationship update is sent (1 log record)
			name:         "different type relationship works with prefixes",
			inputFile:    "relationship/different-types-relationship/01-with-prefixes/input.yaml",
			expectedFile: "relationship/different-types-relationship/01-with-prefixes/expected-output.yaml",
			configFile:   "relationship/different-types-relationship/01-with-prefixes/config.yaml",
		},
		{
			// ~~~ DESCRIPTION ~~~
			// config.yaml
			// 		does not contain source_prefix & destination_prefix
			// 		due to this, there should be match on entity attributes and relationship attributes (3 log records)
			// input.yaml
			//  	attributes are not prefixed
			// expected-output.yaml
			// 		two entity updates are sent, and one relationship update is sent (3 log records)
			name:         "different type relationship works without prefixes",
			inputFile:    "relationship/different-types-relationship/02-without-prefixes/input.yaml",
			expectedFile: "relationship/different-types-relationship/02-without-prefixes/expected-output.yaml",
			configFile:   "relationship/different-types-relationship/02-without-prefixes/config.yaml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var cfg *Config
			var err error
			if tc.configFile != "" {
				cfg, err = loadConfigFromFile(t, filepath.Join("testdata", "metricsToLogs", tc.configFile))
			} else {
				cfg, err = loadConfigFromFile(t, "testdata/config.yaml")
			}
			require.NoError(t, err)

			factory := NewFactory()
			sink := &consumertest.LogsSink{}
			conn, err := factory.CreateMetricsToLogs(context.Background(),
				connectortest.NewNopSettings(metadata.Type), cfg, sink)
			require.NoError(t, err)
			require.NotNil(t, conn)

			require.NoError(t, conn.Start(context.Background(), componenttest.NewNopHost()))
			defer func() {
				assert.NoError(t, conn.Shutdown(context.Background()))
			}()

			testMetrics, err := golden.ReadMetrics(filepath.Join("testdata", "metricsToLogs", tc.inputFile))
			assert.NoError(t, err)
			assert.NoError(t, conn.ConsumeMetrics(context.Background(), testMetrics))

			allLogs := sink.AllLogs()
			if len(tc.expectedFile) == 0 {
				assert.Len(t, allLogs, 0)
				return
			}

			expected, err := golden.ReadLogs(filepath.Join("testdata", "metricsToLogs", tc.expectedFile))

			assert.NoError(t, err)
			assert.Equal(t, allLogs[0].LogRecordCount(), expected.LogRecordCount())
			assert.NoError(t, plogtest.CompareLogs(expected, allLogs[0], plogtest.IgnoreObservedTimestamp()))
		})
	}
}

func loadConfigFromFile(t *testing.T, path string) (*Config, error) {
	t.Helper()

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
