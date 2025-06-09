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
		// ~~~~~~ ENTITIES TESTS ~~~~~~
		{
			// Checks creation of 2 Snowflake entities without any conditions.
			name:   "when config has no conditions, entity is inferred  and log event is sent",
			folder: "entity/01-no-conditions",
		},
		{
			name:   "when log for entity has valid complex condition, log event is sent",
			folder: "entity/02-condition-met",
		},
		{
			// Input is sending insufficient attributes for entity creation
			// No event should be sent.
			name:   "when entity is not inferred no log is sent",
			folder: "entity/03-no-match",
		},
		{
			// Attributes for schema.entities are sufficient for entity creation, but
			// condition in schema.event.entities is not satisfied.
			name:   "when log for entity has not satisfied the condition, no log event is sent",
			folder: "entity/04-condition-not-met",
		},
		// ~~~~~~ SAME TYPE RELATIONSHIP TESTS ~~~~~~
		{
			// Checks that same type relationship for AWS EC2 is sent, together with the two AWS EC2 entities.
			// Uses simple ["true"] conditions.
			// Uses prefixes as all same type relationship tests.
			name:   "when relationship for same type is inferred log event is sent",
			folder: "relationship/same-type-relationship/01-no-conditions",
		},
		{
			// Testing that more complex conditions get evaluated correctly.
			// Checks that events for entities and the relationship are still sent.
			name:   "when same type relationship has valid advanced condition, log event is sent",
			folder: "relationship/same-type-relationship/02-advanced-conditions",
		},
		{
			// Tests that when two entities share an attribute key and value for some required attribute,
			// the events for entities and their relationship are still detected and sent.
			name:   "when relationship for same type and having common id attributes is inferred log event is sent",
			folder: "relationship/same-type-relationship/03-common-attr",
		},
		{
			// Checks that if one of the attributes for the relationship is not set, the relationship is not sent, but is for entities.
			name:   "when relationship for same type is not inferred no log is sent",
			folder: "relationship/same-type-relationship/04-no-match",
		},
		{
			// Checks that when additional attributes are set on the relationship, they are sent with the relationship.
			name:   "when log for same type relationship, log event is sent with relationship attributes",
			folder: "relationship/same-type-relationship/05-extra-attr",
		},
		// ~~~~~~ DIFFERENT TYPE RELATIONSHIP TESTS ~~~~~~
		{
			// Checks that different type relationship supports optional prefixes for source and destination attributes.
			// Since no unprefixed attributes are sent that would match the entities, only relationship log record is sent.
			name:   "different type relationship works with prefixes",
			folder: "relationship/different-types-relationship/01-with-prefixes",
		},
		{
			// Checks that different type relationship works without prefixes configuration.
			// Since there are no prefixes, the attributes match the entities and their relationship, so 3 events are sent.
			name:   "different type relationship works without prefixes",
			folder: "relationship/different-types-relationship/02-without-prefixes",
		},
		{
			// When config.yaml has 2 entities to infer with relationship between them, but one of them
			// is missing required id attribute, relationship log event is not sent and only entity log event is sent for the one entity that was found.
			name:   "when log for different type relationship hasn't all necessary id attributes, log event is sent",
			folder: "relationship/different-types-relationship/03-missing-attr",
		},
		{
			// Relationship condition is not satisfied, so no relationship log event is sent, but entities are.
			name:   "when log for different type relationship has not satisfied the condition, no log relationship event is sent",
			folder: "relationship/different-types-relationship/04-condition-not-met",
		},
		{
			// Checks that when there is an extra attribute, that has nothing to do with entities or relationship,
			// relationship and entities are still sent.
			name:   "when log for different type relationship has redundant attributes, log event is sent",
			folder: "relationship/different-types-relationship/05-redundant-attr",
		},
		{
			// Checks that when relationship condition is satisfied, relationship log event is sent, and the entities also.
			name:   "when log for different type relationship has satisfied the condition, log relationship event is sent",
			folder: "relationship/different-types-relationship/06-condition-met",
		},
		{
			// Relationship should be sent with the extra attributes, and also 2 entity log events.
			name:   "when log for different type relationship, log event is sent with relationship attributes",
			folder: "relationship/different-types-relationship/07-extra-attr",
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
			assert.Equal(t, expected.LogRecordCount(), allLogs[0].LogRecordCount())
			assert.NoError(t, plogtest.CompareLogs(expected, allLogs[0], plogtest.IgnoreObservedTimestamp()))
		})
	}
}

func TestMetricsToLogs(t *testing.T) {
	testCases := []struct {
		name   string
		folder string
	}{
		//  ~~~~~~ ENTITIES TESTS ~~~~~~
		{
			// Checks creation of 2 Snowflake entities without any conditions.
			name:   "when config has no conditions, entity is inferred  and log event is sent",
			folder: "entity/01-no-conditions",
		},
		{
			name:   "when log for entity has satisfied a complex condition, log event is sent",
			folder: "entity/02-condition-met",
		},
		{
			// Input is sending insufficient attributes for entity creation
			name:   "when entity is not inferred no log is sent",
			folder: "entity/03-no-match",
		},
		{
			// Attributes for schema.entities are sufficient for entity creation, but
			// condition in schema.event.entities is not satisfied.
			name:   "when log for entity has not satisfied the condition, no log event is sent",
			folder: "entity/04-condition-not-met",
		},
		//  ~~~~~~ SAME TYPE RELATIONSHIP TESTS ~~~~~~
		{
			// Checks that same type relationship for AWS EC2 is sent, together with the two AWS EC2 entities.
			// Uses simple ["true"] conditions.
			// Uses prefixes as all same type relationship tests.
			name:   "when relationship for same type is inferred log event is sent",
			folder: "relationship/same-type-relationship/01-no-conditions",
		},
		{
			// Testing that more complex conditions get evaluated correctly.
			// Checks that events for entities and relationship are still sent.
			name:   "when same type relationship has valid advanced condition, log event is sent",
			folder: "relationship/same-type-relationship/02-advanced-conditions",
		},
		{
			// Tests that when two entities share an attribute key and value for some required attribute,
			// the events for entities and their relationship are still detected and sent.
			name:   "when relationship for same type and having common id attributes is inferred log event is sent",
			folder: "relationship/same-type-relationship/03-common-attr",
		},
		{
			// Checks that if one of the attributes for the relationship is not set, the relationship is not sent, but is for entities.
			name:   "when relationship for same type is not inferred no log is sent",
			folder: "relationship/same-type-relationship/04-no-match",
		},
		{

			// Checks that when additional attributes are set on the relationship, they are sent with the relationship.
			name:   "when log for same type relationship, log event is sent with relationship attributes",
			folder: "relationship/same-type-relationship/05-extra-attr",
		},
		// ~~~~~~ DIFFERENT TYPE RELATIONSHIP TESTS ~~~~~~
		{
			// Checks that different type relationship supports optional prefixes for source and destination attributes.
			// Since no unprefixed attributes are sent that would match the entities, only relationship log record is sent.
			name:   "different type relationship works with prefixes",
			folder: "relationship/different-types-relationship/01-with-prefixes",
		},
		{
			// Checks that different type relationship works without prefixes configuration.
			// Since there are no prefixes, the attributes match the entities and their relationship, so 3 events are sent.
			name:   "different type relationship works without prefixes",
			folder: "relationship/different-types-relationship/02-without-prefixes",
		},
		{
			// When config.yaml has 2 entities to infer with relationship between them, but one of them
			// is missing required id attribute, relationship log event is not sent and only entity log event is sent for the one entity that was found.
			name:   "when log for different type relationship hasn't all necessary id attributes, log event is sent",
			folder: "relationship/different-types-relationship/03-missing-attr",
		},
		{
			// Relationship condition is not satisfied, so no relationship log event is sent, but entities are.
			name:   "when log for different type relationship has not satisfied the condition, no log relationship event is sent",
			folder: "relationship/different-types-relationship/04-condition-not-met",
		},
		{
			// Checks that when there is an extra attribute, that has nothing to do with entities or relationship,
			// relationship and entities are still sent if they have what is needed.
			name:   "when log for different type relationship has redundant attributes, log event is sent",
			folder: "relationship/different-types-relationship/05-redundant-attr",
		},
		{
			// Checks that when relationship condition is satisfied, relationship log event is sent, and the entities also.
			name:   "when log for different type relationship has satisfied the condition, log relationship event is sent",
			folder: "relationship/different-types-relationship/06-condition-met",
		},
		{
			// Relationship should be sent with the extra attributes, and also 2 entity log events.
			name:   "when log for different type relationship, log event is sent with relationship attributes",
			folder: "relationship/different-types-relationship/07-extra-attr",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := loadConfigFromFile(t, filepath.Join("testdata", "metricsToLogs", tc.folder, "config.yaml"))
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

			inputFile := filepath.Join("testdata", "metricsToLogs", tc.folder, "input.yaml")
			testMetrics, err := golden.ReadMetrics(inputFile)

			assert.NoError(t, err)
			assert.NoError(t, conn.ConsumeMetrics(context.Background(), testMetrics))

			allLogs := sink.AllLogs()
			expectedFile := filepath.Join("testdata", "metricsToLogs", tc.folder, "expected-output.yaml")

			if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
				assert.Len(t, allLogs, 0)
				return
			}

			expected, err := golden.ReadLogs(expectedFile)

			assert.NoError(t, err)
			assert.Equal(t, expected.LogRecordCount(), allLogs[0].LogRecordCount())
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
