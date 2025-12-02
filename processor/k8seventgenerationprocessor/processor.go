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

package k8seventgenerationprocessor

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/k8seventgenerationprocessor/internal/manifests"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

const (
	k8sObjectKind = "k8s.object.kind"
)

type k8seventgenerationprocessor struct {
	cfg               component.Config
	telemetrySettings component.TelemetrySettings
	logger            *zap.Logger
}

type result struct {
	Manifest  any // Can be *manifests.PodManifest, *manifests.EndpointManifest, or *manifests.EndpointSliceManifest
	Timestamp pcommon.Timestamp
}

// processLogs goes through all log records and parse information about Container entities
// and relations based on Endpoints and EndpointSlices from them.
// The entities/relations are created based on all log records from all scope and resource logs.
// The resulting state logs are appended as a new ResourceLogs to the plog.Logs structure that is processed at the time.
func (cp *k8seventgenerationprocessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	resourceLogs := ld.ResourceLogs()
	mCh := make(chan result)
	errCh := make(chan error)

	entityStateEvents := plog.NewLogRecordSlice()
	serviceMappingsLogSlice := plog.NewLogRecordSlice()

	var wg sync.WaitGroup
	wg.Go(func() { cp.generateManifests(mCh, errCh, resourceLogs) })
	wg.Go(func() {
		cp.generateLogRecords(mCh, entityStateEvents, serviceMappingsLogSlice)
	})

	if err, ok := <-errCh; ok {
		return ld, err
	}

	wg.Wait()

	if entityStateEvents.Len() > 0 {
		addEntityStateEventResourceLog(ld, entityStateEvents)
	}

	if serviceMappingsLogSlice.Len() > 0 {
		addServiceMappingsResourceLog(ld, serviceMappingsLogSlice)
	}

	return ld, nil
}

// generateLogRecords appends all LogRecords containing container information to the provided LogRecordSlice.
func (cp *k8seventgenerationprocessor) generateLogRecords(resCh <-chan result, entityStateEvents plog.LogRecordSlice, lrsServiceMappings plog.LogRecordSlice) {
	for res := range resCh {
		switch m := res.Manifest.(type) {
		case *manifests.PodManifest:
			manifestContainers := m.GetContainers()
			cp.logger.Error("manifestContainers", zap.Any("containers", manifestContainers))
			containers := transformContainersToContainerLogs(manifestContainers, m.Metadata, res.Timestamp)
			containers.MoveAndAppendTo(entityStateEvents)
			containerImages := transformContainersToContainerImageLogs(manifestContainers, res.Timestamp)
			containerImages.MoveAndAppendTo(entityStateEvents)
			containerImageRelations := transformContainersToContainerImageRelationsLogs(manifestContainers, m.Metadata, res.Timestamp)
			containerImageRelations.MoveAndAppendTo(entityStateEvents)
		case manifests.ServiceMapping:
			mappings := transformManifestToServiceMappingLogs(m, res.Timestamp)
			mappings.MoveAndAppendTo(lrsServiceMappings)
		}
	}
}

// generateManifests extracts and parses manifests from log records that have k8s.object.kind set to "Pod".
func (cp *k8seventgenerationprocessor) generateManifests(resCh chan<- result, errCh chan<- error, resourceLogs plog.ResourceLogsSlice) {
	defer close(resCh)
	defer close(errCh)

	for _, rl := range resourceLogs.All() {
		scopeLogs := rl.ScopeLogs()

		for _, sl := range scopeLogs.All() {
			logRecords := sl.LogRecords()

			for _, lr := range logRecords.All() {
				manifest, err := extractManifest(lr)

				if err != nil {
					cp.logger.Error("Error while unmarshaling manifest", zap.Error(err))
					errCh <- err
					return
				} else if manifest != nil {
					timestamp := getTimestamp(lr)
					resCh <- result{
						Manifest:  manifest,
						Timestamp: timestamp,
					}
				}
			}
		}
	}
}

func extractManifest(lr plog.LogRecord) (manifestPointer any, err error) {
	attrs := lr.Attributes()
	kind, ok := attrs.Get(k8sObjectKind)
	if !ok {
		return nil, nil // Not a k8s manifest
	}

	objType := kind.Str()
	var manifest any

	switch objType {
	case "Pod":
		var m manifests.PodManifest
		body := lr.Body().AsString()
		err = json.Unmarshal([]byte(body), &m)
		manifest = &m
	case "Endpoints":
		var m manifests.EndpointManifest
		body := lr.Body().AsString()
		err = json.Unmarshal([]byte(body), &m)
		manifest = &m
	case "EndpointSlice":
		var m manifests.EndpointSliceManifest
		body := lr.Body().AsString()
		err = json.Unmarshal([]byte(body), &m)
		manifest = &m
	default:
		return nil, nil // Not a supported k8s manifest
	}

	if err != nil {
		return nil, err
	}

	return manifest, nil
}

// getTimestamp returns the timestamp of the log record.
// If observed timestamp is set, it is returned, otherwise the timestamp or current time is returned.
func getTimestamp(lr plog.LogRecord) pcommon.Timestamp {
	if !lr.ObservedTimestamp().AsTime().IsZero() {
		return lr.ObservedTimestamp()
	}

	if !lr.Timestamp().AsTime().IsZero() {
		return lr.Timestamp()
	}

	return pcommon.NewTimestampFromTime(time.Now())
}

func (cp *k8seventgenerationprocessor) Start(_ context.Context, _ component.Host) error {
	cp.logger.Info("Starting container processor")
	return nil
}

func (cp *k8seventgenerationprocessor) Shutdown(_ context.Context) error {
	cp.logger.Info("Shutting down container processor")
	return nil
}
