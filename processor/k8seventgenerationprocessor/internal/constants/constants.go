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

package constants

const (
	// Entity Event Types
	AttributeOtelEntityEventType  = "otel.entity.event.type"
	AttributeOtelEntityType       = "otel.entity.type"
	AttributeOtelEntityID         = "otel.entity.id"
	AttributeOtelEntityAttributes = "otel.entity.attributes"

	AttributeOtelEntityRelationshipType                = "otel.entity_relationship.type"
	AttributeOtelEntityRelationshipSourceEntityID      = "otel.entity_relationship.source_entity.id"
	AttributeOtelEntityRelationshipDestinationEntityID = "otel.entity_relationship.destination_entity.id"
	AttributeOtelEntityRelationshipAttributes          = "otel.entity_relationship.attributes"

	EventTypeEntityState             = "entity_state"
	EventTypeEntityRelationshipState = "entity_relationship_state"

	EntityTypeVulnerability                = "VulnerabilityDetail"
	EntityTypeKubernetesDeployment         = "KubernetesDeployment"
	EntityTypeKubernetesDaemonSet          = "KubernetesDaemonSet"
	EntityTypeKubernetesStatefulSet        = "KubernetesStatefulSet"
	RelationshipTypeVulnerabilityFinding   = "VulnerabilityFinding"
	AttributeOtelEntityRelationshipSrcType = "otel.entity_relationship.source_entity.type"
	AttributeOtelEntityRelationshipDstType = "otel.entity_relationship.destination_entity.type"

	// Vulnerability Attributes
	AttributeVulnerabilityID          = "vulnerability.id"
	AttributeVulnerabilityDescription = "vulnerability.description"
	AttributeVulnerabilitySeverity    = "vulnerability.severity"
	AttributeVulnerabilityScoreBase   = "vulnerability.score.base"
	AttributeVulnerabilityEnumeration = "vulnerability.enumeration"
	AttributeVulnerabilityReference   = "vulnerability.reference"
	AttributeCweID                    = "cwe.id"

	// Scanner Attributes
	AttributeScannerVendor  = "scannerVendor"
	AttributeScannerName    = "scannerName"
	AttributeScannerVersion = "scannerVersion"
)
