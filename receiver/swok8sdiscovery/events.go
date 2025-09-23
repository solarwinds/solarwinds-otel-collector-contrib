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
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

// databaseEvent is an internal representation before converting to logs.
type databaseEvent struct {
	DatabaseType string
	Namespace    string
	Endpoint     string  // service name if available, else pod name
	Ports        []int32 // chosen ports (default preferred)
	WorkloadKind string  // Deployment/StatefulSet/DaemonSet/Job/CronJob
	WorkloadName string
}

const (
	clusterUidEnv                = "CLUSTER_UID"
	discoveredDatabaseEntityType = "DiscoveredDatabaseInstance"
	entityState                  = "entity_state"
	relationshipState            = "entity_relationship_state"

	// Attributes for OTel entity events identification
	otelEntityEventAsLog              = "otel.entity.event_as_log"
	otelEntityEventType               = "otel.entity.event.type"
	otelEntityRelationType            = "otel.entity_relationship.type"
	otelEntityRelationSourceType      = "otel.entity_relationship.source_entity.type"
	otelEntityRelationSourceID        = "otel.entity_relationship.source_entity.id"
	otelEntityRelationDestinationType = "otel.entity_relationship.destination_entity.type"
	otelEntityRelationDestinationID   = "otel.entity_relationship.destination_entity.id"

	swEntityType         = "otel.entity.type"
	k8sNamespace         = "k8s.namespace.name"
	swDiscoveryDbName    = "sw.discovery.dbo.name"
	swDiscoveryDbAddress = "sw.discovery.dbo.address"
	swDiscoveryDbType    = "sw.discovery.dbo.type"
	swDiscoveryId        = "sw.discovery.id"
	swDiscoverySource    = "sw.discovery.source"

	// Attributes for telemetry mapping
	otelEntityId         = "otel.entity.id"
	otelEntityAttributes = "otel.entity.attributes"
	swK8sClusterUid      = "sw.k8s.cluster.uid"
)

// publishDatabaseEvent publishes structured log record for database discovery outcome.
func (r *swok8sdiscoveryReceiver) publishDatabaseEvent(ctx context.Context, discoveryId string, ev databaseEvent) {
	logs := plog.NewLogs()
	scopeLogs := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty()
	scopeLogs.Scope().Attributes().PutBool(otelEntityEventAsLog, true)

	logRecord := scopeLogs.LogRecords().AppendEmpty()
	attrs := logRecord.Attributes()

	// compose database name
	name := ev.Endpoint
	if ev.WorkloadName != "" {
		name += "#" + ev.WorkloadName
	}

	// compose full endpoint with ports
	address := ev.Endpoint
	if len(ev.Ports) > 0 {
		address += ":" + strings.Join(portsAsStrings(ev.Ports), ",")
	}

	attrs.PutStr(otelEntityEventType, entityState)
	attrs.PutStr(swEntityType, discoveredDatabaseEntityType)
	attrs.PutStr(swDiscoverySource, r.config.Reporter)
	// add workload attributes fo filtering
	if ev.WorkloadKind != "" {
		attrs.PutStr("k8s."+strings.ToLower(ev.WorkloadKind)+".name", ev.WorkloadName)
		attrs.PutStr(k8sNamespace, ev.Namespace)
		attrs.PutStr(swK8sClusterUid, r.clusterUid)
	}

	keys := attrs.PutEmptyMap(otelEntityId)
	keys.PutStr(swDiscoveryDbAddress, address)
	keys.PutStr(swDiscoveryDbType, ev.DatabaseType)
	keys.PutStr(swDiscoveryId, discoveryId)

	optional := attrs.PutEmptyMap(otelEntityAttributes)
	optional.PutStr(swDiscoveryDbName, name)

	r.consumer.ConsumeLogs(ctx, logs)

	r.publishRelationShip(ctx, ev, keys)
}

func (r *swok8sdiscoveryReceiver) publishRelationShip(ctx context.Context, ev databaseEvent, dbKeys pcommon.Map) {

	if ev.WorkloadKind == "" || ev.WorkloadName == "" {
		return
	}

	logs := plog.NewLogs()
	scopeLogs := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty()
	scopeLogs.Scope().Attributes().PutBool(otelEntityEventAsLog, true)

	logRecord := scopeLogs.LogRecords().AppendEmpty()
	attrs := logRecord.Attributes()

	attrs.PutStr(otelEntityEventType, relationshipState)
	attrs.PutStr(otelEntityRelationType, "DiscoveredBy")

	//
	// source entity
	attrs.PutStr(otelEntityRelationSourceType, "Kubernetes"+ev.WorkloadKind)
	src_ids := attrs.PutEmptyMap(otelEntityRelationSourceID)
	src_ids.PutStr("k8s."+strings.ToLower(ev.WorkloadKind)+".name", ev.WorkloadName)
	src_ids.PutStr(k8sNamespace, ev.Namespace)
	src_ids.PutStr(swK8sClusterUid, r.clusterUid)

	//
	// destination entity
	attrs.PutStr(otelEntityRelationDestinationType, discoveredDatabaseEntityType)
	dst_ids := attrs.PutEmptyMap(otelEntityRelationDestinationID)
	dbKeys.CopyTo(dst_ids)

	r.consumer.ConsumeLogs(ctx, logs)
}

func portsAsStrings(ports []int32) []string {
	if len(ports) == 0 {
		return nil
	}
	res := make([]string, len(ports))
	for i, p := range ports {
		res[i] = strconv.FormatInt(int64(p), 10)
	}
	return res
}
