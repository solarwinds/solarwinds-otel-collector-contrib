package internal

import (
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

type ContainerInfo struct {
	ContainerID       string
	IsRunInContainerd bool
}

const (
	cloudProviderAttribute = "cloud.provider"
	hostIDAttribute        = "host.id"
	hostNameAttribute      = "host.name"
	osTypeAttribute        = "os.type"

	clientIdAttribute = "sw.uams.client.id"
)

func (h *ContainerInfo) addHostAttributes(
	resourceAttributes pcommon.Map,
	configuredAttributes map[string]string,
) {

	_, cloudProviderExists := resourceAttributes.Get(cloudProviderAttribute)
	clientId, clientIdExists := configuredAttributes[clientIdAttribute]

	if clientIdExists && clientId != "" {
		// Replace host ID attribute with client ID only for hosts that are not cloud-based.
		// Cloud-based hosts have unique cloud instance host ID that is historically used
		// in the system.
		if !cloudProviderExists {
			resourceAttributes.PutStr(hostIDAttribute, clientId)
		}

		// Replace host ID attribute with client ID if the OTel plugin is running inside a container.
		// Some containers running on cloud-based hosts can have the same host ID leading to problems
		// with differentiation of watched containers, so it has to be replaced with client ID which is unique.
		if cloudProviderExists && h.ContainerID != "" {
			resourceAttributes.PutStr(hostIDAttribute, clientId)
		}
	}

	// Replace host name attribute with container ID only for containerd containers on AWS machines
	// to match host name reported by UAMS Client and host name reported by OTel plugin.
	if cloudProviderExists && h.IsRunInContainerd {
		resourceAttributes.PutStr(hostNameAttribute, h.ContainerID)
	}

	// Unify the os.type values, supported are only Windows and Linux for now.
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
