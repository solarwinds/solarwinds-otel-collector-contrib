package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestAddPluginAttributes_HostIdScenario1_NonCloudHost(t *testing.T) {
	// Scenario 1: Non-cloud host - hostId should be set to clientId
	pp := PluginProperties{
		OverrideHostnameEnv:  "",
		ContainerHostnameEnv: "",
		IsInContainerd:       false,
		ContainerID:          "",
		ReceiverName:         "",
		ReceiverDisplayName:  "",
		ClientID:             "test-client-id",
	}

	attributes := pcommon.NewMap()
	// No cloud.provider attribute = non-cloud host

	pp.addPluginAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-client-id", hostId.Str())
}

func TestAddPluginAttributes_HostIdScenario2_CloudHostWithContainer(t *testing.T) {
	// Scenario 2: Cloud host with container - hostId should be set to clientId
	pp := PluginProperties{
		OverrideHostnameEnv:  "",
		ContainerHostnameEnv: "",
		IsInContainerd:       false,
		ContainerID:          "container-123",
		ReceiverName:         "",
		ReceiverDisplayName:  "",
		ClientID:             "test-client-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	pp.addPluginAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-client-id", hostId.Str())
}

func TestAddPluginAttributes_HostIdScenario3_GcpHost(t *testing.T) {
	// Scenario 3: GCP host - hostId should be projectId:zoneId:instanceId
	pp := PluginProperties{
		OverrideHostnameEnv:  "",
		ContainerHostnameEnv: "",
		IsInContainerd:       false,
		ContainerID:          "",
		ReceiverName:         "",
		ReceiverDisplayName:  "",
		ClientID:             "test-client-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "gcp")
	attributes.PutStr("cloud.account.id", "my-project")
	attributes.PutStr("cloud.availability_zone", "us-central1-a")
	attributes.PutStr("host.id", "instance-123")

	pp.addPluginAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "my-project:us-central1-a:instance-123", hostId.Str())
}

func TestAddPluginAttributes_GcpHostMissingAttributes(t *testing.T) {
	// GCP host with missing attributes - hostId should be removed
	pp := PluginProperties{
		OverrideHostnameEnv:  "",
		ContainerHostnameEnv: "",
		IsInContainerd:       false,
		ContainerID:          "",
		ReceiverName:         "",
		ReceiverDisplayName:  "",
		ClientID:             "test-client-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "gcp")
	attributes.PutStr("host.id", "instance-123")
	// Missing cloud.account.id and cloud.availability_zone

	pp.addPluginAttributes(attributes)

	_, exists := attributes.Get("host.id")
	require.False(t, exists)
}

func TestAddPluginAttributes_CloudHostWithoutContainer(t *testing.T) {
	// Cloud host without container - hostId should remain unchanged
	pp := PluginProperties{
		OverrideHostnameEnv:  "",
		ContainerHostnameEnv: "",
		IsInContainerd:       false,
		ContainerID:          "", // No container
		ReceiverName:         "",
		ReceiverDisplayName:  "",
		ClientID:             "test-client-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	pp.addPluginAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "original-host-id", hostId.Str())
}

func TestAddPluginAttributes_HostnameScenarios(t *testing.T) {
	testCases := map[string]struct {
		isInContainerd           bool
		containerHostnameEnv     string
		overrideHostnameEnv      string
		cloudProvider            string
		containerID              string
		expectedHostname         string
		expectedHostnameOverride string
	}{
		"containerd with cloud provider": {
			isInContainerd:   true,
			cloudProvider:    "aws",
			containerID:      "container-123",
			expectedHostname: "container-123",
		},
		"container hostname env set": {
			containerHostnameEnv: "custom-hostname",
			expectedHostname:     "custom-hostname",
		},
		"override hostname env set": {
			overrideHostnameEnv:      "override-hostname",
			expectedHostnameOverride: "override-hostname",
		},
		"no special hostname handling": {
			// Should not modify hostname
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			attributes := pcommon.NewMap()
			if tc.cloudProvider != "" {
				attributes.PutStr("cloud.provider", tc.cloudProvider)
			}
			attributes.PutStr("host.name", "original-hostname")

			pp := PluginProperties{
				OverrideHostnameEnv:  tc.overrideHostnameEnv,
				ContainerHostnameEnv: tc.containerHostnameEnv,
				IsInContainerd:       tc.isInContainerd,
				ContainerID:          tc.containerID,
				ReceiverName:         "",
				ReceiverDisplayName:  "",
				ClientID:             "test-client-id",
			}
			pp.addPluginAttributes(attributes)

			if tc.expectedHostname != "" {
				hostname, exists := attributes.Get("host.name")
				require.True(t, exists)
				require.Equal(t, tc.expectedHostname, hostname.Str())
			}

			if tc.expectedHostnameOverride != "" {
				hostnameOverride, exists := attributes.Get("host.name.override")
				require.True(t, exists)
				require.Equal(t, tc.expectedHostnameOverride, hostnameOverride.Str())
			}
		})
	}
}

func TestAddPluginAttributes_OsTypeNormalization(t *testing.T) {
	pp := PluginProperties{
		OverrideHostnameEnv:  "",
		ContainerHostnameEnv: "",
		IsInContainerd:       false,
		ContainerID:          "",
		ReceiverName:         "",
		ReceiverDisplayName:  "",
		ClientID:             "test-client-id",
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

			pp.addPluginAttributes(attributes)

			osType, exists := attributes.Get("os.type")
			require.True(t, exists)
			require.Equal(t, expected, osType.Str())
		})
	}
}

func TestAddPluginAttributes_ReceiverAttributes(t *testing.T) {
	pp := PluginProperties{
		OverrideHostnameEnv:  "",
		ContainerHostnameEnv: "",
		IsInContainerd:       false,
		ContainerID:          "",
		ReceiverName:         "test-receiver",
		ReceiverDisplayName:  "Test Receiver",
		ClientID:             "test-client-id",
	}
	attributes := pcommon.NewMap()

	pp.addPluginAttributes(attributes)

	receiverName, exists := attributes.Get("sw.uams.receiver.name")
	require.True(t, exists)
	require.Equal(t, "test-receiver", receiverName.Str())

	receiverDisplayName, exists := attributes.Get("sw.uams.receiver.display_name")
	require.True(t, exists)
	require.Equal(t, "Test Receiver", receiverDisplayName.Str())
}

func TestAddPluginAttributes_DoesNotOverwriteExistingReceiverName(t *testing.T) {
	pp := PluginProperties{
		OverrideHostnameEnv:  "",
		ContainerHostnameEnv: "",
		IsInContainerd:       false,
		ContainerID:          "",
		ReceiverName:         "new-receiver",
		ReceiverDisplayName:  "",
		ClientID:             "test-client-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("sw.uams.receiver.name", "existing-receiver")

	pp.addPluginAttributes(attributes)

	receiverName, exists := attributes.Get("sw.uams.receiver.name")
	require.True(t, exists)
	require.Equal(t, "existing-receiver", receiverName.Str())

	_, exists = attributes.Get("sw.uams.receiver.display_name")
	require.False(t, exists)
}

func TestAddPluginAttributes_OverwriteExistingReceiverDisplayName(t *testing.T) {
	pp := PluginProperties{
		OverrideHostnameEnv:  "",
		ContainerHostnameEnv: "",
		IsInContainerd:       false,
		ContainerID:          "",
		ReceiverName:         "new-receiver",
		ReceiverDisplayName:  "New Display Name",
		ClientID:             "test-client-id",
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("sw.uams.receiver.display_name", "DISPLAY NAME WILL BE OVERWRITTEN")

	pp.addPluginAttributes(attributes)

	receiverDisplayName, exists := attributes.Get("sw.uams.receiver.display_name")
	require.True(t, exists)
	require.Equal(t, "New Display Name", receiverDisplayName.Str())
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
