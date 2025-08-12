package internal

import (
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

type PluginProperties struct {
	IsInContainerd bool   `mapstructure:"is.in.containerd"`
	ContainerID    string `mapstructure:"container.id"`
	ClientID       string `mapstructure:"client.id"`
}

const (
	cloudProviderAttribute = "cloud.provider"
	hostIDAttribute        = "host.id"
	hostNameAttribute      = "host.name"
	osTypeAttribute        = "os.type"
)

func (h *PluginProperties) addPluginAttributes(
	resourceAttributes pcommon.Map,
) {

	// telemetry
	_, cloudProviderExists := resourceAttributes.Get(cloudProviderAttribute)

	// Replace host ID attribute with client ID only for hosts that are not cloud-based.
	// Cloud-based hosts have unique cloud instance host ID that is historically used
	// in the system.
	if !cloudProviderExists {
		resourceAttributes.PutStr(hostIDAttribute, h.ClientID)
	}

	// telemetry + poll
	// Replace host ID attribute with client ID if the OTel plugin is running inside a container.
	// Some containers running on cloud-based hosts can have the same host ID leading to problems
	// with differentiation of watched containers, so it has to be replaced with client ID which is unique.
	if cloudProviderExists && h.ContainerID != "" {
		resourceAttributes.PutStr(hostIDAttribute, h.ClientID)
	}

	// telemetry + poll
	// Replace host name attribute with container ID only for containerd containers on AWS machines
	// to match host name reported by UAMS Client and host name reported by OTel plugin.
	if cloudProviderExists && h.IsInContainerd {
		resourceAttributes.PutStr(hostNameAttribute, h.ContainerID)
	}

	// telemetry
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
