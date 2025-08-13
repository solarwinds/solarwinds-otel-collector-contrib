package internal

import (
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

type HostAttributes struct {
	ContainerID       string
	IsRunInContainerd bool
	ClientId          string
}

const (
	cloudProviderAttribute = "cloud.provider"
	hostIDAttribute        = "host.id"
	hostNameAttribute      = "host.name"
	osTypeAttribute        = "os.type"
)

func (h *HostAttributes) ApplyAttributes(
	resourceAttributes pcommon.Map,
) {

	cloudProvider, cloudProviderExists := resourceAttributes.Get(cloudProviderAttribute)

	if h.ClientId != "" {
		// Replace host ID attribute with client ID only for hosts that are not cloud-based.
		// Cloud-based hosts have unique cloud instance host ID that is historically used
		// in the system.
		if !cloudProviderExists {
			resourceAttributes.PutStr(hostIDAttribute, h.ClientId)
		}

		// Replace host ID attribute with client ID if the OTel plugin is running inside a container.
		// Some containers running on cloud-based hosts can have the same host ID leading to problems
		// with differentiation of watched containers, so it has to be replaced with client ID which is unique.
		if cloudProviderExists && h.ContainerID != "" {
			resourceAttributes.PutStr(hostIDAttribute, h.ClientId)
		}
	}

	// Replace host name attribute with container ID only for containerd containers on AWS machines
	// to match host name reported by different monitoring components.
	if cloudProvider.Str() == "aws" && h.IsRunInContainerd {
		resourceAttributes.PutStr(hostNameAttribute, h.ContainerID)
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
