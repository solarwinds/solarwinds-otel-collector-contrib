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
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

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
		name   string
		folder string
	}{
		// ~~~~~~ ENTITIES TESTS ~~~~~~
		{
			name:   "when log for entity has valid complex condition, log event is sent",
			folder: "entity/condition-met",
		},
		{
			// Attributes for schema.entities are sufficient for entity creation, but
			// condition in schema.event.entities is not satisfied.
			name:   "when log for entity has not satisfied the condition, no log event is sent",
			folder: "entity/condition-not-met",
		},
		{
			name:   "when action is set as delete, delete log event is sent",
			folder: "entity/delete-action",
		},
		{
			// Checks creation of 2 Snowflake entities without any conditions.
			name:   "when config has no conditions, entity is inferred  and log event is sent",
			folder: "entity/no-conditions",
		},
		{
			// Input is sending insufficient attributes for entity creation
			// No event should be sent.
			name:   "when entity is not inferred no log is sent",
			folder: "entity/no-match",
		},
		// ~~~~~~ SAME TYPE RELATIONSHIP TESTS ~~~~~~
		{
			// Testing that more complex conditions get evaluated correctly.
			// Checks that events for entities and the relationship are still sent.
			name:   "when same type relationship has valid advanced condition, log event is sent",
			folder: "relationship/same-type-relationship/advanced-conditions",
		},
		{
			// Tests that when two entities share an attribute key and value for some required attribute,
			// the events for entities and their relationship are still detected and sent.
			name:   "when relationship for same type and having common id attributes is inferred log event is sent",
			folder: "relationship/same-type-relationship/common-attr",
		},
		{
			name:   "when action is set as delete, delete log event is sent",
			folder: "relationship/same-type-relationship/delete-action",
		},
		{
			// Checks that when additional attributes are set on the relationship, they are sent with the relationship.
			name:   "when log for same type relationship, log event is sent with relationship attributes",
			folder: "relationship/same-type-relationship/extra-attr",
		},
		{
			// Checks that same type relationship for AWS EC2 is sent, together with the two AWS EC2 entities.
			// Uses simple ["true"] conditions.
			// Uses prefixes as all same type relationship tests.
			name:   "when relationship for same type is inferred log event is sent",
			folder: "relationship/same-type-relationship/no-conditions",
		},
		{
			// Checks that if one of the attributes for the relationship is not set, the relationship is not sent, but is for entities.
			name:   "when relationship for same type is not inferred no log is sent",
			folder: "relationship/same-type-relationship/no-match",
		},
		// ~~~~~~ DIFFERENT TYPE RELATIONSHIP TESTS ~~~~~~
		{
			// Checks that when relationship condition is satisfied, relationship log event is sent, and the entities also.
			name:   "when log for different type relationship has satisfied the condition, log relationship event is sent",
			folder: "relationship/different-types-relationship/condition-met",
		},
		{
			// Relationship condition is not satisfied, so no relationship log event is sent, but entities are.
			name:   "when log for different type relationship has not satisfied the condition, no log relationship event is sent",
			folder: "relationship/different-types-relationship/condition-not-met",
		},
		{
			name:   "when action is set as delete, delete log event is sent",
			folder: "relationship/different-types-relationship/delete-action",
		},
		{
			// Relationship should be sent with the extra attributes, and also 2 entity log events.
			name:   "when log for different type relationship, log event is sent with relationship attributes",
			folder: "relationship/different-types-relationship/extra-attr",
		},
		{
			// When config.yaml has 2 entities to infer with relationship between them, but one of them
			// is missing required id attribute, relationship log event is not sent and only entity log event is sent for the one entity that was found.
			name:   "when log for different type relationship hasn't all necessary id attributes, log event is sent",
			folder: "relationship/different-types-relationship/missing-attr",
		},
		{
			// Checks that when there is an extra attribute, that has nothing to do with entities or relationship,
			// relationship and entities are still sent.
			name:   "when log for different type relationship has redundant attributes, log event is sent",
			folder: "relationship/different-types-relationship/redundant-attr",
		},
		{
			// Checks that different type relationship supports optional prefixes for source and destination attributes.
			// Since no unprefixed attributes are sent that would match the entities, only relationship log record is sent.
			name:   "different type relationship works with prefixes",
			folder: "relationship/different-types-relationship/with-prefixes",
		},
		{
			// Checks that different type relationship works without prefixes configuration.
			// Since there are no prefixes, the attributes match the entities and their relationship, so 3 events are sent.
			name:   "different type relationship works without prefixes",
			folder: "relationship/different-types-relationship/without-prefixes",
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
			name:   "when log for entity has satisfied a complex condition, log event is sent",
			folder: "entity/condition-met",
		},
		{
			// Attributes for schema.entities are sufficient for entity creation, but
			// condition in schema.event.entities is not satisfied.
			name:   "when log for entity has not satisfied the condition, no log event is sent",
			folder: "entity/condition-not-met",
		},
		{
			name:   "when action is set as delete, delete log event is sent",
			folder: "entity/delete-action",
		},
		{
			// Checks creation of 2 Snowflake entities without any conditions.
			name:   "when config has no conditions, entity is inferred  and log event is sent",
			folder: "entity/no-conditions",
		},
		{
			// Input is sending insufficient attributes for entity creation
			name:   "when entity is not inferred no log is sent",
			folder: "entity/no-match",
		},
		//  ~~~~~~ SAME TYPE RELATIONSHIP TESTS ~~~~~~
		{
			// Testing that more complex conditions get evaluated correctly.
			// Checks that events for entities and relationship are still sent.
			name:   "when same type relationship has valid advanced condition, log event is sent",
			folder: "relationship/same-type-relationship/advanced-conditions",
		},
		{
			// Tests that when two entities share an attribute key and value for some required attribute,
			// the events for entities and their relationship are still detected and sent.
			name:   "when relationship for same type and having common id attributes is inferred log event is sent",
			folder: "relationship/same-type-relationship/common-attr",
		},
		{
			name:   "when action is set as delete, delete log event is sent",
			folder: "relationship/same-type-relationship/delete-action",
		},
		{

			// Checks that when additional attributes are set on the relationship, they are sent with the relationship.
			name:   "when log for same type relationship, log event is sent with relationship attributes",
			folder: "relationship/same-type-relationship/extra-attr",
		},
		{
			// Checks that same type relationship for AWS EC2 is sent, together with the two AWS EC2 entities.
			// Uses simple ["true"] conditions.
			// Uses prefixes as all same type relationship tests.
			name:   "when relationship for same type is inferred log event is sent",
			folder: "relationship/same-type-relationship/no-conditions",
		},
		{
			// Checks that if one of the attributes for the relationship is not set, the relationship is not sent, but is for entities.
			name:   "when relationship for same type is not inferred no log is sent",
			folder: "relationship/same-type-relationship/no-match",
		},
		// ~~~~~~ DIFFERENT TYPE RELATIONSHIP TESTS ~~~~~~
		{
			// Checks that when relationship condition is satisfied, relationship log event is sent, and the entities also.
			name:   "when log for different type relationship has satisfied the condition, log relationship event is sent",
			folder: "relationship/different-types-relationship/condition-met",
		},
		{
			// Relationship condition is not satisfied, so no relationship log event is sent, but entities are.
			name:   "when log for different type relationship has not satisfied the condition, no log relationship event is sent",
			folder: "relationship/different-types-relationship/condition-not-met",
		},
		{
			name:   "when action is set as delete, delete log event is sent",
			folder: "relationship/different-types-relationship/delete-action",
		},
		{
			// Relationship should be sent with the extra attributes, and also 2 entity log events.
			name:   "when log for different type relationship, log event is sent with relationship attributes",
			folder: "relationship/different-types-relationship/extra-attr",
		},
		{
			// When config.yaml has 2 entities to infer with relationship between them, but one of them
			// is missing required id attribute, relationship log event is not sent and only entity log event is sent for the one entity that was found.
			name:   "when log for different type relationship hasn't all necessary id attributes, log event is sent",
			folder: "relationship/different-types-relationship/missing-attr",
		},
		{
			// Checks that when there is an extra attribute, that has nothing to do with entities or relationship,
			// relationship and entities are still sent if they have what is needed.
			name:   "when log for different type relationship has redundant attributes, log event is sent",
			folder: "relationship/different-types-relationship/redundant-attr",
		},
		{
			// Checks that different type relationship supports optional prefixes for source and destination attributes.
			// Since no unprefixed attributes are sent that would match the entities, only relationship log record is sent.
			name:   "different type relationship works with prefixes",
			folder: "relationship/different-types-relationship/with-prefixes",
		},
		{
			// Checks that different type relationship works without prefixes configuration.
			// Since there are no prefixes, the attributes match the entities and their relationship, so 3 events are sent.
			name:   "different type relationship works without prefixes",
			folder: "relationship/different-types-relationship/without-prefixes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFolder := filepath.Join("testdata", "metricsToLogs", tc.folder)
			cfg, err := loadConfigFromFile(t, filepath.Join(testFolder, "config.yaml"))
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

			inputFile := filepath.Join(testFolder, "input.yaml")
			testMetrics, err := golden.ReadMetrics(inputFile)

			assert.NoError(t, err)
			assert.NoError(t, conn.ConsumeMetrics(context.Background(), testMetrics))

			allLogs := sink.AllLogs()
			expectedFile := filepath.Join(testFolder, "expected-output.yaml")

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

// Using cache.
// Sending relationship update first to populate cache, then
// sending delete action, should send delete log event.
func TestRelationshipDeleteWithCache(t *testing.T) {
	testFolder := filepath.Join("testdata", "metricsToLogs", "relationship/different-types-relationship/delete-action-cached")
	cfg, err := loadConfigFromFile(t, filepath.Join(testFolder, "config.yaml"))
	require.NoError(t, err)
	factory := NewFactory()
	sink := &consumertest.LogsSink{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conn, err := factory.CreateMetricsToLogs(ctx,
		connectortest.NewNopSettings(metadata.Type), cfg, sink)
	require.NoError(t, err)
	require.NotNil(t, conn)

	require.NoError(t, conn.Start(ctx, componenttest.NewNopHost()))
	defer func() {
		assert.NoError(t, conn.Shutdown(ctx))
	}()

	// 1st incoming log, relationship update
	inputFile := filepath.Join(testFolder, "input1.yaml")
	testMetrics, err := golden.ReadMetrics(inputFile)
	assert.NoError(t, err)
	assert.NoError(t, conn.ConsumeMetrics(ctx, testMetrics))

	allLogs := sink.AllLogs()
	expectedFile := filepath.Join(testFolder, "expected-output1.yaml")

	expected, err := golden.ReadLogs(expectedFile)
	assert.NoError(t, err)
	assert.Equal(t, expected.LogRecordCount(), allLogs[0].LogRecordCount())
	assert.NoError(t, plogtest.CompareLogs(expected, allLogs[0], plogtest.IgnoreObservedTimestamp()))

	// 2nd incoming log, relationship delete
	inputFile2 := filepath.Join(testFolder, "input2.yaml")
	sink.Reset()
	testMetrics2, err := golden.ReadMetrics(inputFile2)
	assert.NoError(t, err)
	assert.NoError(t, conn.ConsumeMetrics(ctx, testMetrics2))
	allLogs2 := sink.AllLogs()

	expectedFile2 := filepath.Join(testFolder, "expected-output2.yaml")
	expected2, err := golden.ReadLogs(expectedFile2)
	assert.NoError(t, err)
	assert.Equal(t, expected.LogRecordCount(), allLogs2[0].LogRecordCount())
	assert.NoError(t, plogtest.CompareLogs(expected2, allLogs2[0], plogtest.IgnoreObservedTimestamp()))

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

// Test that the connector consumes a log or metric from which
// it infers a relationship, produces a relationship event and
// then waits for the cache expiration to ensure that the relationship delete event is produced.
func TestRelationshipCacheExpiration(t *testing.T) {
	testFolder := filepath.Join("testdata", "logsToLogs", "relationship", "cacheExpiration")
	cfg, err := loadConfigFromFile(t, filepath.Join(testFolder, "config.yaml"))
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

	// Consume input logs or metrics that infer a relationship
	inputFile := filepath.Join(testFolder, "input.yaml")
	testLogs, err := golden.ReadLogs(inputFile)
	require.NoError(t, err)
	require.NoError(t, conn.ConsumeLogs(context.Background(), testLogs))

	// Wait for the cache expiration to ensure that the relationship delete event is produced
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	timeout := time.After(10 * time.Second)
	for sink.LogRecordCount() < 4 { // 2 entities, 1 relationship and 1 delete event
		select {
		case <-timeout:
			require.Fail(t, "timed out waiting for logs to be processed")
		case <-ticker.C:
			fmt.Printf("Waiting for logs to be processed...\n")
		}
	}

	allLogs := sink.AllLogs()

	// Check that the delete relationship event is produced
	expectedFile := filepath.Join(testFolder, "expected-output.yaml")
	expected, err := golden.ReadLogs(expectedFile)
	require.NoError(t, err)
	assert.Equal(t, expected.LogRecordCount(), allLogs[1].LogRecordCount())
	assert.NoError(t, plogtest.CompareLogs(expected, allLogs[1], plogtest.IgnoreObservedTimestamp()))
}
