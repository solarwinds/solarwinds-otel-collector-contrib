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
	"fmt"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

type attrUpdaterFunc func(pcommon.Map)

func watchObjectsToLogData(event *watch.Event, observedAt time.Time, config *K8sObjectsConfig, attrUpdater func(pcommon.Map)) (plog.Logs, error) {
	udata, ok := event.Object.(*unstructured.Unstructured)
	if !ok {
		return plog.Logs{}, fmt.Errorf("received data that wasnt unstructure, %v", event)
	}

	ul := unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{{
			Object: map[string]any{
				"type":   string(event.Type),
				"object": udata.Object,
			},
		}},
	}

	if attrUpdater == nil {
		attrUpdater = func(attrs pcommon.Map) {}
	}

	return unstructuredListToLogData(&ul, observedAt, config, func(attrs pcommon.Map) {
		objectMeta := udata.Object["metadata"].(map[string]any)
		name := objectMeta["name"].(string)
		if name != "" {
			attrs.PutStr("event.domain", "k8s")
			attrs.PutStr("event.name", name)
		}
	}, attrUpdater), nil
}

func pullObjectsToLogData(event *unstructured.UnstructuredList, observedAt time.Time, config *K8sObjectsConfig) plog.Logs {
	return unstructuredListToLogData(event, observedAt, config)
}

func unstructuredListToLogData(event *unstructured.UnstructuredList, observedAt time.Time, config *K8sObjectsConfig, attrUpdaters ...attrUpdaterFunc) plog.Logs {
	out := plog.NewLogs()
	resourceLogs := out.ResourceLogs()
	namespaceResourceMap := make(map[string]plog.LogRecordSlice)

	for _, e := range event.Items {
		logSlice, ok := namespaceResourceMap[getNamespace(e)]
		if !ok {
			rl := resourceLogs.AppendEmpty()
			resourceAttrs := rl.Resource().Attributes()
			if namespace := getNamespace(e); namespace != "" {
				resourceAttrs.PutStr(string(semconv.K8SNamespaceNameKey), namespace)
			}
			sl := rl.ScopeLogs().AppendEmpty()
			logSlice = sl.LogRecords()
			namespaceResourceMap[getNamespace(e)] = logSlice
		}
		record := logSlice.AppendEmpty()
		record.SetObservedTimestamp(pcommon.NewTimestampFromTime(observedAt))

		attrs := record.Attributes()
		attrs.PutStr("k8s.resource.name", config.gvr.Resource)

		for _, attrUpdate := range attrUpdaters {
			attrUpdate(attrs)
		}

		dest := record.Body()
		destMap := dest.SetEmptyMap()
		//nolint:errcheck
		destMap.FromRaw(e.Object)
	}
	return out
}

func getNamespace(e unstructured.Unstructured) string {
	// first, try to use the GetNamespace() method, which checks for the metadata.namespace property
	if namespace := e.GetNamespace(); namespace != "" {
		return namespace
	}
	// try to look up namespace in object.metadata.namespace (for objects reported via watch mode)
	if namespace, ok, _ := unstructured.NestedString(e.Object, "object", "metadata", "namespace"); ok {
		return namespace
	}
	return ""
}
