package solarwindsprocessor

import (
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
)

type PluginProperties struct {
	OverrideHostnameEnv  string `mapstructure:"override.hostname.env"`
	ContainerHostnameEnv string `mapstructure:"container.hostname.env"`
	IsInContainerd       bool   `mapstructure:"is.in.containerd"`
	ContainerID          string `mapstructure:"container.id"`
	ReceiverName         string `mapstructure:"receiver.name"`
	ReceiverDisplayName  string `mapstructure:"receiver.display_name"`
	ClientID             string `mapstructure:"client.id"`
}

const (
	cloudProviderAttribute       = "cloud.provider"
	hostIDAttribute              = "host.id"
	hostNameAttribute            = "host.name"
	hostNameOverrideAttribute    = "host.name.override"
	receiverNameAttribute        = "sw.uams.receiver.name"
	receiverDisplayNameAttribute = "sw.uams.receiver.display_name"
	osTypeAttribute              = "os.type"

	// Google Cloud Platform attributes.
	cloudAccountID        = "cloud.account.id"
	cloudAvailabilityZone = "cloud.availability_zone"
)

func (h *PluginProperties) addPluginAttributes(
	resourceAttributes pcommon.Map,
) {
	// Set host name override attribute with overrideHostnameEnv. This env var can be set system-wide and is set
	// to match host name reported by UAMS Client and host name reported by OTel plugin.
	if h.OverrideHostnameEnv != "" {
		resourceAttributes.PutStr(hostNameOverrideAttribute, h.OverrideHostnameEnv)
	}

	cloudProvider, cloudProviderExists := resourceAttributes.Get(cloudProviderAttribute)

	// Replace host ID attribute with client ID only for hosts that are not cloud-based.
	// Cloud-based hosts have unique cloud instance host ID that is historically used
	// in the system.
	if !cloudProviderExists {
		resourceAttributes.PutStr(hostIDAttribute, h.ClientID)
	}

	// Replace host ID attribute with client ID if the OTel plugin is running inside a container.
	// Some containers running on cloud-based hosts can have the same host ID leading to problems
	// with differentiation of watched containers, so it has to be replaced with client ID which is unique.
	if cloudProviderExists && h.ContainerID != "" {
		resourceAttributes.PutStr(hostIDAttribute, h.ClientID)
	}

	// If the cloud provider is GPC, we need to set the host ID attribute to a combination of
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

	// Replace host name attribute with container ID only for containerd containers on AWS machines
	// to match host name reported by UAMS Client and host name reported by OTel plugin.
	if cloudProviderExists && h.IsInContainerd {
		resourceAttributes.PutStr(hostNameAttribute, h.ContainerID)
	}

	// Replace host name attribute with containerHostnameEnv for all containers if UAMS_CONTAINER_HOSTNAME is set
	// to match host name reported by UAMS Client and host name reported by OTel plugin.
	if h.ContainerHostnameEnv != "" {
		resourceAttributes.PutStr(hostNameAttribute, h.ContainerHostnameEnv)
	}

	// Unify the os.type values, supported are only Windows and Linux for now.
	// Everything what's not Windows should be identified as Linux.
	if osTypeValue, exists := resourceAttributes.Get(osTypeAttribute); exists {
		normalizeOsType(osTypeValue)
	}

	// Don't overwrite receiver name if it's already set (e.g. filelog sets receiver name to hostmetrics).
	if _, exists := resourceAttributes.Get(receiverNameAttribute); !exists {
		resourceAttributes.PutStr(receiverNameAttribute, h.ReceiverName)
	}

	if h.ReceiverDisplayName != "" {
		resourceAttributes.PutStr(receiverDisplayNameAttribute, h.ReceiverDisplayName)
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
