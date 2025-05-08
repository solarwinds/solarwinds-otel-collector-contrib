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
	Manifest  Manifest
	Timestamp pcommon.Timestamp
}

// processLogs go through all log records and parse information about containers from them.
// Containers are created based on all log records from all scope and resource logs.
// Containers related logs are appended as a new ResourceLogs to the plog.Logs structure that is processed at the time.
func (cp *k8seventgenerationprocessor) processLogs(_ context.Context, ld plog.Logs) (plog.Logs, error) {
	resourceLogs := ld.ResourceLogs()
	mCh := make(chan result)
	errCh := make(chan error)

	logSlice := plog.NewLogRecordSlice()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go cp.generateLogRecords(mCh, wg, logSlice)
	go cp.generateManifests(mCh, errCh, wg, resourceLogs)

	select {
	case err, ok := <-errCh:
		if !ok {
			break
		}
		return ld, err
	}

	wg.Wait()

	if logSlice.Len() > 0 {
		rl := addContainersResourceLog(ld)
		lrs := rl.ScopeLogs().At(0).LogRecords()
		logSlice.CopyTo(lrs)
	}

	return ld, nil
}

// generateLogRecords appends all LogRecords containing container information to the provided LogRecordSlice.
func (cp *k8seventgenerationprocessor) generateLogRecords(resCh <-chan result, wg *sync.WaitGroup, lrs plog.LogRecordSlice) {
	defer wg.Done()
	for res := range resCh {
		containers := transformManifestToContainerLogs(res.Manifest, res.Timestamp)
		for i := range containers.Len() {
			lr := containers.At(i)
			lr.CopyTo(lrs.AppendEmpty())
		}
	}
}

// generateManifests extracts and parses manifests from log records that have k8s.object.kind set to "Pod".
func (cp *k8seventgenerationprocessor) generateManifests(resCh chan<- result, errCh chan<- error, wg *sync.WaitGroup, resourceLogs plog.ResourceLogsSlice) {
	defer wg.Done()
	defer close(resCh)
	defer close(errCh)

	for i := range resourceLogs.Len() {
		rl := resourceLogs.At(i)
		scopeLogs := rl.ScopeLogs()

		for j := range scopeLogs.Len() {
			sl := scopeLogs.At(j)
			logRecords := sl.LogRecords()

			for k := range logRecords.Len() {
				lr := logRecords.At(k)
				attrs := lr.Attributes()

				// processor is interested in pods only, since containers are related to pods
				if !isPodLog(attrs) {
					continue
				}
				body := lr.Body().AsString()
				var m Manifest

				err := json.Unmarshal([]byte(body), &m)
				if err != nil {
					cp.logger.Error("Error while unmarshalling manifest", zap.Error(err))
					errCh <- err
					return
				} else {
					timestamp := getTimestamp(lr)
					resCh <- result{
						Manifest:  m,
						Timestamp: timestamp,
					}
				}
			}
		}
	}
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

func isPodLog(attributes pcommon.Map) bool {
	kind, _ := attributes.Get(k8sObjectKind)
	return kind.Str() == "Pod"
}

func (cp *k8seventgenerationprocessor) Start(_ context.Context, _ component.Host) error {
	cp.logger.Info("Starting container processor")
	return nil
}

func (cp *k8seventgenerationprocessor) Shutdown(_ context.Context) error {
	cp.logger.Info("Shutting down container processor")
	return nil
}
