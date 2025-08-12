package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestAddPluginAttributes_HostIdScenario1_NonCloudHost(t *testing.T) {
	// Scenario 1: Non-cloud host - hostId should be set to clientId
	pp := ContainerInfo{
		IsRunInContainerd: false,
		ContainerID:       "",
	}

	configuredAttributes := map[string]string{
		"sw.uams.client.id": "test-client-id",
	}

	attributes := pcommon.NewMap()
	// No cloud.provider attribute = non-cloud host

	pp.addHostAttributes(attributes, configuredAttributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-client-id", hostId.Str())
}

func TestAddPluginAttributes_HostIdScenario2_CloudHostWithContainer(t *testing.T) {
	// Scenario 2: Cloud host with container - hostId should be set to clientId
	pp := ContainerInfo{
		IsRunInContainerd: false,
		ContainerID:       "container-123",
	}

	configuredAttributes := map[string]string{
		"sw.uams.client.id": "test-client-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	pp.addHostAttributes(attributes, configuredAttributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-client-id", hostId.Str())
}

func TestAddPluginAttributes_CloudHostWithoutContainer(t *testing.T) {
	// Cloud host without container - hostId should remain unchanged
	pp := ContainerInfo{
		IsRunInContainerd: false,
		ContainerID:       "", // No container
	}
	configuredAttributes := map[string]string{
		"sw.uams.client.id": "test-client-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	pp.addHostAttributes(attributes, configuredAttributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "original-host-id", hostId.Str())
}

func TestAddPluginAttributes_HostnameScenarios(t *testing.T) {
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
		"containerd with GCP cloud provider": {
			isRunInContainerd: true,
			cloudProvider:     "gcp",
			containerID:       "container-123",
			expectedHostname:  "container-123",
		},
		"without containerd with GCP cloud provider": {
			isRunInContainerd: true,
			cloudProvider:     "gcp",
			containerID:       "container-123",
			expectedHostname:  "",
		},
		"no special hostname handling": {
			// Should not modify hostname
		},
	}

	configuredAttributes := map[string]string{
		"sw.uams.client.id": "test-client-id",
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			attributes := pcommon.NewMap()
			if tc.cloudProvider != "" {
				attributes.PutStr("cloud.provider", tc.cloudProvider)
			}
			attributes.PutStr("host.name", "original-hostname")

			pp := ContainerInfo{
				IsRunInContainerd: tc.isRunInContainerd,
				ContainerID:       tc.containerID,
			}
			pp.addHostAttributes(attributes, configuredAttributes)

			if tc.expectedHostname != "" {
				hostname, exists := attributes.Get("host.name")
				require.True(t, exists)
				require.Equal(t, tc.expectedHostname, hostname.Str())
			}
		})
	}
}

func TestAddPluginAttributes_OsTypeNormalization(t *testing.T) {
	pp := ContainerInfo{
		IsRunInContainerd: false,
		ContainerID:       "",
	}

	testCases := map[string]string{
		"windows": "Windows",
		"linux":   "Linux",
		"unix":    "Linux",
		"darwin":  "Linux",
	}

	configuredAttributes := map[string]string{
		"sw.uams.client.id": "test-client-id",
	}

	for input, expected := range testCases {
		t.Run(input, func(t *testing.T) {
			attributes := pcommon.NewMap()
			attributes.PutStr("os.type", input)

			pp.addHostAttributes(attributes, configuredAttributes)

			osType, exists := attributes.Get("os.type")
			require.True(t, exists)
			require.Equal(t, expected, osType.Str())
		})
	}
}

func Test_OsTypeIsNormalized(t *testing.T) {
	testCases := map[string]struct {
		Input    pcommon.Value
		Expected string
	}{
		"windows is Windows": {
			Input:    pcommon.NewValueStr("windows"),
			Expected: "Windows",
		},
		"linux is Linux": {
			Input:    pcommon.NewValueStr("linux"),
			Expected: "Linux",
		},
		"unix is Linux": {
			Input:    pcommon.NewValueStr("unix"),
			Expected: "Linux",
		},
		"other string value is Linux": {
			Input:    pcommon.NewValueStr("something completely different"),
			Expected: "Linux",
		},
		"non-string value is Linux": {
			Input:    pcommon.NewValueInt(42),
			Expected: "Linux",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			osTypeValue := testCase.Input
			normalizeOsType(osTypeValue)

			require.Equal(t, testCase.Expected, osTypeValue.Str())
		})
	}
}
