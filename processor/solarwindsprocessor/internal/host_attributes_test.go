package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestApplyAttributes_HostId_NonCloudHostWithFallbackHostId(t *testing.T) {
	pp := HostAttributes{
		IsRunInContainerd: false,
		ContainerID:       "",
		FallbackHostID:    "test-fallback-host-id",
	}

	attributes := pcommon.NewMap()

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-fallback-host-id", hostId.Str())
}

func TestApplyAttributes_HostId_NonCloudHostWithoutFallbackHostId(t *testing.T) {
	pp := HostAttributes{
		IsRunInContainerd: false,
		ContainerID:       "",
		FallbackHostID:    "",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr(hostIDAttribute, "telemetry-host-id")

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "", hostId.Str()) // Should be empty string from FallbackHostID, not original
}

func TestApplyAttributes_HostId_CloudHostWithContainer(t *testing.T) {
	pp := HostAttributes{
		IsRunInContainerd: false,
		ContainerID:       "container-123",
		FallbackHostID:    "test-fallback-host-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-fallback-host-id", hostId.Str())
}

func TestApplyAttributes_HostId_CloudHostWithoutContainer(t *testing.T) {
	// Cloud host without container - hostId should remain unchanged
	pp := HostAttributes{
		IsRunInContainerd: false,
		ContainerID:       "", // No container
		FallbackHostID:    "test-fallback-host-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "original-host-id", hostId.Str())
}

func TestApplyAttributes_HostId_BiosUuidVsFallbackHostIdScenarios(t *testing.T) {
	testCases := map[string]struct {
		biosUuid       string
		biosUuidExists bool
		fallbackHostID string
		cloudProvider  string
		containerID    string
		expectedHostId string
		description    string
	}{
		"bios uuid from telemetry exists and not empty - non cloud": {
			biosUuid:       "12345678-1234-1234-1234-123456789abc",
			biosUuidExists: true,
			fallbackHostID: "fallback-id-from-config",
			cloudProvider:  "",
			containerID:    "",
			expectedHostId: "12345678-1234-1234-1234-123456789abc",
			description:    "Should use BIOS UUID from telemetry when available on non-cloud host",
		},
		"bios uuid from telemetry exists but empty - non cloud": {
			biosUuid:       "",
			biosUuidExists: true,
			fallbackHostID: "fallback-id-from-config",
			cloudProvider:  "",
			containerID:    "",
			expectedHostId: "fallback-id-from-config",
			description:    "Should use fallback host ID from config when BIOS UUID from telemetry is empty on non-cloud host",
		},
		"bios uuid from telemetry does not exist - non cloud": {
			biosUuid:       "",
			biosUuidExists: false,
			fallbackHostID: "fallback-id-from-config",
			cloudProvider:  "",
			containerID:    "",
			expectedHostId: "fallback-id-from-config",
			description:    "Should use fallback host ID from config when BIOS UUID doesn't exist in telemetry on non-cloud host",
		},
		"bios uuid from telemetry exists - cloud host without container": {
			biosUuid:       "12345678-1234-1234-1234-123456789abc",
			biosUuidExists: true,
			fallbackHostID: "fallback-id-from-config",
			cloudProvider:  "aws",
			containerID:    "",
			expectedHostId: "original-cloud-host-id",
			description:    "Should preserve original host ID on cloud host without container, ignoring both BIOS UUID and fallback host ID",
		},
		"bios uuid from telemetry exists - cloud host with container": {
			biosUuid:       "12345678-1234-1234-1234-123456789abc",
			biosUuidExists: true,
			fallbackHostID: "fallback-id-from-config",
			cloudProvider:  "aws",
			containerID:    "container-123",
			expectedHostId: "12345678-1234-1234-1234-123456789abc",
			description:    "Should use BIOS UUID from telemetry on cloud host with container",
		},
		"no bios uuid from telemetry - cloud host with container": {
			biosUuid:       "",
			biosUuidExists: false,
			fallbackHostID: "fallback-id-from-config",
			cloudProvider:  "aws",
			containerID:    "container-123",
			expectedHostId: "fallback-id-from-config",
			description:    "Should use fallback host ID from config on cloud host with container when no BIOS UUID in telemetry",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			pp := HostAttributes{
				IsRunInContainerd: false,
				ContainerID:       tc.containerID,
				FallbackHostID:    tc.fallbackHostID,
			}

			attributes := pcommon.NewMap()
			if tc.cloudProvider != "" {
				attributes.PutStr("cloud.provider", tc.cloudProvider)
				attributes.PutStr("host.id", "original-cloud-host-id")
			}
			if tc.biosUuidExists {
				attributes.PutStr("host.bios-uuid", tc.biosUuid)
			}

			pp.ApplyAttributes(attributes)

			hostId, exists := attributes.Get("host.id")
			require.True(t, exists, tc.description)
			require.Equal(t, tc.expectedHostId, hostId.Str(), tc.description)
		})
	}
}

func TestApplyAttributes_HostnameScenarios_CombinationsOfCloudProviderAndContainerd(t *testing.T) {
	testCases := map[string]struct {
		isRunInContainerd bool
		containerID       string
		cloudProvider     string
		originalHostname  string
		expectedHostname  string
		shouldCheckResult bool
	}{
		"containerd with AWS cloud provider": {
			isRunInContainerd: true,
			cloudProvider:     "aws",
			containerID:       "container-123",
			originalHostname:  "original-hostname",
			expectedHostname:  "container-123",
			shouldCheckResult: true,
		},
		"no containerd with AWS cloud provider": {
			isRunInContainerd: false,
			cloudProvider:     "aws",
			containerID:       "container-123",
			originalHostname:  "original-hostname",
			expectedHostname:  "original-hostname",
			shouldCheckResult: true,
		},
		"containerd with GCP cloud provider": {
			isRunInContainerd: true,
			cloudProvider:     "gcp",
			containerID:       "container-123",
			originalHostname:  "original-hostname",
			expectedHostname:  "original-hostname",
			shouldCheckResult: true,
		},
		"no containerd with GCP cloud provider": {
			isRunInContainerd: false,
			cloudProvider:     "gcp",
			containerID:       "container-123",
			originalHostname:  "original-hostname",
			expectedHostname:  "original-hostname",
			shouldCheckResult: true,
		},
		"no special hostname handling": {
			isRunInContainerd: false,
			cloudProvider:     "",
			containerID:       "container-123",
			originalHostname:  "original-hostname",
			expectedHostname:  "original-hostname",
			shouldCheckResult: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			attributes := pcommon.NewMap()
			if tc.cloudProvider != "" {
				attributes.PutStr("cloud.provider", tc.cloudProvider)
			}
			attributes.PutStr("host.name", tc.originalHostname)

			pp := HostAttributes{
				IsRunInContainerd: tc.isRunInContainerd,
				ContainerID:       tc.containerID,
				FallbackHostID:    "test-client-id",
			}
			pp.ApplyAttributes(attributes)

			if tc.shouldCheckResult {
				hostname, exists := attributes.Get("host.name")
				require.True(t, exists)
				require.Equal(t, tc.expectedHostname, hostname.Str())
			}
		})
	}
}

func TestApplyAttributes_OsType_IsNormalized(t *testing.T) {
	pp := HostAttributes{
		IsRunInContainerd: false,
		ContainerID:       "",
		FallbackHostID:    "test-fallback-host-id",
	}

	testCases := map[string]string{
		"windows": "Windows",
		"linux":   "Linux",
		"unix":    "Linux",
		"darwin":  "Linux",
	}

	for input, expected := range testCases {
		t.Run(input, func(t *testing.T) {
			attributes := pcommon.NewMap()
			attributes.PutStr("os.type", input)

			pp.ApplyAttributes(attributes)

			osType, exists := attributes.Get("os.type")
			require.True(t, exists)
			require.Equal(t, expected, osType.Str())
		})
	}
}
