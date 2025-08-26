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

package internal

import (
	"strings"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/pkg/container"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

type HostAttributesDecorator struct {
	ContainerID       string
	IsRunInContainerd bool
	FallbackHostID    string
}

const (
	cloudProviderAttribute = "cloud.provider"
	hostIDAttribute        = "host.id"
	hostBiosUuid           = "host.bios-uuid"
	hostNameAttribute      = "host.name"
	osTypeAttribute        = "os.type"

	// Google Cloud Platform attributes.
	cloudAccountID        = "cloud.account.id"
	cloudAvailabilityZone = "cloud.availability_zone"
)

func NewHostAttributes(hd HostDecoration, logger *zap.Logger) (*HostAttributesDecorator, error) {
	cp := container.NewProvider(logger)

	instanceId, err := cp.ReadContainerInstanceID()
	if err != nil {
		return nil, err
	}

	return &HostAttributesDecorator{
		ContainerID:       instanceId,
		IsRunInContainerd: cp.IsRunInContainerd(),
		FallbackHostID:    hd.FallbackHostID,
	}, nil
}

func (h *HostAttributesDecorator) ApplyAttributes(
	resourceAttributes pcommon.Map,
) {
	if h == nil {
		return
	}

	cloudProvider, cloudProviderExists := resourceAttributes.Get(cloudProviderAttribute)
	var hostId string
	biosUuid, biosUuidExists := resourceAttributes.Get(hostBiosUuid)
	if biosUuidExists && biosUuid.Str() != "" {
		hostId = biosUuid.Str()
	} else {
		hostId = h.FallbackHostID
	}

	// Replace host ID attribute with client ID only for hosts that are not cloud-based.
	// Cloud-based hosts have unique cloud instance host ID that is historically used
	// in the system.
	if !cloudProviderExists {
		resourceAttributes.PutStr(hostIDAttribute, hostId)
	}

	// Replace host ID attribute with client ID if the OTel plugin is running inside a container.
	// Some containers running on cloud-based hosts can have the same host ID leading to problems
	// with differentiation of watched containers, so it has to be replaced with client ID which is unique.
	if cloudProviderExists && h.ContainerID != "" {
		resourceAttributes.PutStr(hostIDAttribute, hostId)
	}

	// Replace host name attribute with container ID only for containerd containers on AWS machines
	// to match host name reported by different monitoring components.
	if cloudProvider.Str() == "aws" && h.IsRunInContainerd {
		resourceAttributes.PutStr(hostNameAttribute, h.ContainerID)
	}

	// If the cloud provider is GCP, we need to set the host ID attribute to a combination of
	// cloud account ID, availability zone, and instance ID to make sure all GCP hosts have unique host IDs.
	if cloudProvider.Str() == "gcp" {
		hostID := getGcpHostID(resourceAttributes)
		if hostID != "" {
			resourceAttributes.PutStr(hostIDAttribute, hostID)
		} else {
			instanceID, _ := resourceAttributes.Get(hostIDAttribute)
			zap.L().Warn("could not determine GCP host ID", zap.String("instance_id", instanceID.Str()))
			resourceAttributes.Remove(hostIDAttribute)
		}
	}

	// Unify the os.type values, supported are only Windows and Linux.
	// Everything what's not Windows should be identified as Linux.
	if osTypeValue, exists := resourceAttributes.Get(osTypeAttribute); exists {
		normalizeOsType(osTypeValue)
	}
}

func normalizeOsType(osTypeValue pcommon.Value) {
	originalOsType := osTypeValue.Str()
	if strings.EqualFold(originalOsType, "windows") {
		osTypeValue.SetStr("Windows")
	} else {
		osTypeValue.SetStr("Linux")
	}
}

func getGcpHostID(attributes pcommon.Map) string {
	projectID, _ := attributes.Get(cloudAccountID)
	zoneID, _ := attributes.Get(cloudAvailabilityZone)
	instanceID, _ := attributes.Get(hostIDAttribute)

	if projectID.Str() != "" && zoneID.Str() != "" && instanceID.Str() != "" {
		return projectID.Str() + ":" + zoneID.Str() + ":" + instanceID.Str()
	}

	return ""
}
