package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestApplyAttributes_HostId_NonCloudHostWithClientId(t *testing.T) {
	pp := HostAttributes{
		IsRunInContainerd: false,
		ContainerID:       "",
		ClientId:          "test-client-id",
	}

	attributes := pcommon.NewMap()

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-client-id", hostId.Str())
}

func TestApplyAttributes_HostId_NonCloudHostWithoutClientId(t *testing.T) {
	pp := HostAttributes{
		IsRunInContainerd: false,
		ContainerID:       "",
		ClientId:          "",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr(hostIDAttribute, "telemetry-host-id")

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "telemetry-host-id", hostId.Str())
}

func TestApplyAttributes_HostId_CloudHostWithContainer(t *testing.T) {
	pp := HostAttributes{
		IsRunInContainerd: false,
		ContainerID:       "container-123",
		ClientId:          "test-client-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-client-id", hostId.Str())
}

func TestApplyAttributes_HostId_CloudHostWithoutContainer(t *testing.T) {
	// Cloud host without container - hostId should remain unchanged
	pp := HostAttributes{
		IsRunInContainerd: false,
		ContainerID:       "", // No container
		ClientId:          "test-client-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "original-host-id", hostId.Str())
}

func TestApplyAttributes_HostnameScenarios_CombinationsOfCloudProviderAndContainerd(t *testing.T) {
	testCases := map[string]struct {
		isRunInContainerd bool
		containerID       string
		cloudProvider     string
		expectedHostname  string
	}{
		"containerd with AWS cloud provider": {
			isRunInContainerd: true,
			cloudProvider:     "aws",
			containerID:       "container-123",
			expectedHostname:  "container-123",
		},
		"no containerd with AWS cloud provider": {
			isRunInContainerd: false,
			cloudProvider:     "aws",
			containerID:       "container-123",
			expectedHostname:  "original-hostname",
		},
		"containerd with GCP cloud provider": {
			isRunInContainerd: true,
			cloudProvider:     "gcp",
			containerID:       "container-123",
			expectedHostname:  "original-hostname",
		},
		"without containerd with GCP cloud provider": {
			isRunInContainerd: true,
			cloudProvider:     "gcp",
			containerID:       "container-123",
			expectedHostname:  "",
		},
		"no special hostname handling": {
			// Should not modify hostname
			isRunInContainerd: false,
			cloudProvider:     "",
			containerID:       "container-123",
			expectedHostname:  "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			attributes := pcommon.NewMap()
			if tc.cloudProvider != "" {
				attributes.PutStr("cloud.provider", tc.cloudProvider)
			}
			attributes.PutStr("host.name", "original-hostname")

			pp := HostAttributes{
				IsRunInContainerd: tc.isRunInContainerd,
				ContainerID:       tc.containerID,
				ClientId:          "test-client-id",
			}
			pp.ApplyAttributes(attributes)

			if tc.expectedHostname != "" {
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
		ClientId:          "test-client-id",
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
