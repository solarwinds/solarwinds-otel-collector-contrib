package solarwindsprocessor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

// MockIDsProvider for testing
type MockIDsProvider struct {
	clientID    string
	containerID string
	instanceID  uamsid.InstanceID
	pluginID    uamsid.ComponentID
	hasClientID bool
}

func (m *MockIDsProvider) ClientID() string              { return m.clientID }
func (m *MockIDsProvider) ContainerID() string           { return m.containerID }
func (m *MockIDsProvider) InstanceID() uamsid.InstanceID { return m.instanceID }
func (m *MockIDsProvider) PluginID() uamsid.ComponentID  { return m.pluginID }
func (m *MockIDsProvider) HasClientID() bool             { return m.hasClientID }
func (m *MockIDsProvider) Start(ctx context.Context)     {}
func (m *MockIDsProvider) Stop(ctx context.Context)      {}

func createTestProcessor(idsProvider IDsProvider, isInContainerd bool, containerHostnameEnv, overrideHostnameEnv string) *uamsProcessor {
	receiver := meta.Receiver{
		Name:        "test-receiver",
		DisplayName: "Test Receiver",
	}
	return newProcessor(idsProvider, receiver, []Attribute{}, isInContainerd, containerHostnameEnv, overrideHostnameEnv)
}

func TestAddAttributes_HostIdScenario1_NonCloudHost(t *testing.T) {
	// Scenario 1: Non-cloud host - hostId should be set to clientId
	instanceId, _ := uamsid.NewInstanceID("test-instance")
	idsProvider := &MockIDsProvider{
		clientID:    "test-client-id",
		instanceID:  instanceId,
		pluginID:    uamsid.ComponentID{},
		hasClientID: true,
	}
	processor := createTestProcessor(idsProvider, false, "", "")

	attributes := pcommon.NewMap()
	// No cloud.provider attribute = non-cloud host

	processor.addAttributes(context.Background(), attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-client-id", hostId.Str())
}

func TestAddAttributes_HostIdScenario2_CloudHostWithContainer(t *testing.T) {
	// Scenario 2: Cloud host with container - hostId should be set to clientId
	instanceId, _ := uamsid.NewInstanceID("test-instance")
	pluginId, _ := uamsid.NewComponentID("test-plugin")
	idsProvider := &MockIDsProvider{
		clientID:    "test-client-id",
		containerID: "container-123",
		instanceID:  instanceId,
		pluginID:    pluginId,
		hasClientID: true,
	}
	processor := createTestProcessor(idsProvider, false, "", "")

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	processor.addAttributes(context.Background(), attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-client-id", hostId.Str())
}

func TestAddAttributes_HostIdScenario3_GcpHost(t *testing.T) {
	// Scenario 3: GCP host - hostId should be projectId:zoneId:instanceId
	instanceId, _ := uamsid.NewInstanceID("test-instance")
	pluginId, _ := uamsid.NewComponentID("test-plugin")
	idsProvider := &MockIDsProvider{
		clientID:    "test-client-id",
		instanceID:  instanceId,
		pluginID:    pluginId,
		hasClientID: true,
	}
	processor := createTestProcessor(idsProvider, false, "", "")

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "gcp")
	attributes.PutStr("cloud.account.id", "my-project")
	attributes.PutStr("cloud.availability_zone", "us-central1-a")
	attributes.PutStr("host.id", "instance-123")

	processor.addAttributes(context.Background(), attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "my-project:us-central1-a:instance-123", hostId.Str())
}

func TestAddAttributes_GcpHostMissingAttributes(t *testing.T) {
	// GCP host with missing attributes - hostId should be removed
	instanceId, _ := uamsid.NewInstanceID("test-instance")
	pluginId, _ := uamsid.NewComponentID("test-plugin")
	idsProvider := &MockIDsProvider{
		clientID:    "test-client-id",
		instanceID:  instanceId,
		pluginID:    pluginId,
		hasClientID: true,
	}
	processor := createTestProcessor(idsProvider, false, "", "")

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "gcp")
	attributes.PutStr("host.id", "instance-123")
	// Missing cloud.account.id and cloud.availability_zone

	processor.addAttributes(context.Background(), attributes)

	_, exists := attributes.Get("host.id")
	require.False(t, exists)
}

func TestAddAttributes_CloudHostWithoutContainer(t *testing.T) {
	// Cloud host without container - hostId should remain unchanged
	instanceId, _ := uamsid.NewInstanceID("test-instance")
	pluginId, _ := uamsid.NewComponentID("test-plugin")
	idsProvider := &MockIDsProvider{
		clientID:    "test-client-id",
		containerID: "", // No container
		instanceID:  instanceId,
		pluginID:    pluginId,
		hasClientID: true,
	}
	processor := createTestProcessor(idsProvider, false, "", "")

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	processor.addAttributes(context.Background(), attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "original-host-id", hostId.Str())
}

func TestAddAttributes_HostnameScenarios(t *testing.T) {
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
			instanceId, _ := uamsid.NewInstanceID("test-instance")
			pluginId, _ := uamsid.NewComponentID("test-plugin")
			idsProvider := &MockIDsProvider{
				clientID:    "test-client-id",
				containerID: tc.containerID,
				instanceID:  instanceId,
				pluginID:    pluginId,
				hasClientID: true,
			}
			processor := createTestProcessor(idsProvider, tc.isInContainerd, tc.containerHostnameEnv, tc.overrideHostnameEnv)

			attributes := pcommon.NewMap()
			if tc.cloudProvider != "" {
				attributes.PutStr("cloud.provider", tc.cloudProvider)
			}
			attributes.PutStr("host.name", "original-hostname")

			processor.addAttributes(context.Background(), attributes)

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

func TestAddAttributes_OsTypeNormalization(t *testing.T) {
	instanceId, _ := uamsid.NewInstanceID("test-instance")
	pluginId, _ := uamsid.NewComponentID("test-plugin")
	idsProvider := &MockIDsProvider{
		clientID:    "test-client-id",
		instanceID:  instanceId,
		pluginID:    pluginId,
		hasClientID: true,
	}
	processor := createTestProcessor(idsProvider, false, "", "")

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

			processor.addAttributes(context.Background(), attributes)

			osType, exists := attributes.Get("os.type")
			require.True(t, exists)
			require.Equal(t, expected, osType.Str())
		})
	}
}

func TestAddAttributes_CoreAttributes(t *testing.T) {
	instanceID, _ := uamsid.NewInstanceID("test-instance")
	pluginID, _ := uamsid.NewComponentID("test-plugin")
	idsProvider := &MockIDsProvider{
		clientID:    "test-client-id",
		instanceID:  instanceID,
		pluginID:    pluginID,
		hasClientID: true,
	}
	processor := createTestProcessor(idsProvider, false, "", "")

	attributes := pcommon.NewMap()

	processor.addAttributes(context.Background(), attributes)

	// Check core attributes are always set
	clientID, exists := attributes.Get("sw.uams.client.id")
	require.True(t, exists)
	require.Equal(t, "test-client-id", clientID.Str())

	pluginInstanceID, exists := attributes.Get("sw.uams.plugin.instance.id")
	require.True(t, exists)
	require.Equal(t, instanceID.String(), pluginInstanceID.Str())

	pluginIDAttr, exists := attributes.Get("sw.uams.plugin.id")
	require.True(t, exists)
	require.Equal(t, pluginID.String(), pluginIDAttr.Str())
}

func TestAddAttributes_ReceiverAttributes(t *testing.T) {
	instanceId, _ := uamsid.NewInstanceID("test-instance")
	pluginId, _ := uamsid.NewComponentID("test-plugin")
	idsProvider := &MockIDsProvider{
		clientID:    "test-client-id",
		instanceID:  instanceId,
		pluginID:    pluginId,
		hasClientID: true,
	}
	processor := createTestProcessor(idsProvider, false, "", "")

	attributes := pcommon.NewMap()

	processor.addAttributes(context.Background(), attributes)

	receiverName, exists := attributes.Get("sw.uams.receiver.name")
	require.True(t, exists)
	require.Equal(t, "test-receiver", receiverName.Str())

	receiverDisplayName, exists := attributes.Get("sw.uams.receiver.display_name")
	require.True(t, exists)
	require.Equal(t, "Test Receiver", receiverDisplayName.Str())
}

func TestAddAttributes_DoesNotOverwriteExistingReceiverName(t *testing.T) {
	instanceId, _ := uamsid.NewInstanceID("test-instance")
	pluginId, _ := uamsid.NewComponentID("test-plugin")
	idsProvider := &MockIDsProvider{
		clientID:    "test-client-id",
		instanceID:  instanceId,
		pluginID:    pluginId,
		hasClientID: true,
	}
	processor := createTestProcessor(idsProvider, false, "", "")

	attributes := pcommon.NewMap()
	attributes.PutStr("sw.uams.receiver.name", "existing-receiver")

	processor.addAttributes(context.Background(), attributes)

	receiverName, exists := attributes.Get("sw.uams.receiver.name")
	require.True(t, exists)
	require.Equal(t, "existing-receiver", receiverName.Str())
}

func TestAddAttributes_CustomAttributes(t *testing.T) {
	customAttr := NewAttribute("custom.key", "custom-value", func(attributes pcommon.Map, name string, value string) {
		attributes.PutStr(name, value)
	})

	instanceId, _ := uamsid.NewInstanceID("test-instance")
	pluginId, _ := uamsid.NewComponentID("test-plugin")
	idsProvider := &MockIDsProvider{
		clientID:    "test-client-id",
		instanceID:  instanceId,
		pluginID:    pluginId,
		hasClientID: true,
	}
	receiver := meta.Receiver{Name: "test-receiver"}
	processor := newProcessor(idsProvider, receiver, []Attribute{customAttr}, false, "", "")

	attributes := pcommon.NewMap()

	processor.addAttributes(context.Background(), attributes)

	customValue, exists := attributes.Get("custom.key")
	require.True(t, exists)
	require.Equal(t, "custom-value", customValue.Str())
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

func Test_GpcHostIdIfAllAttributesArePresent(t *testing.T) {
	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.account.id", "project-id")
	attributes.PutStr("cloud.availability_zone", "zone-id")
	attributes.PutStr("host.id", "instance-id")

	actualHostID := getGcpHostID(attributes)
	require.Equal(t, "project-id:zone-id:instance-id", actualHostID)
}

func Test_GpcHostIdIfProjectIdIsMissing(t *testing.T) {
	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.availability_zone", "zone-id")
	attributes.PutStr("host.id", "instance-id")

	actualHostID := getGcpHostID(attributes)

	require.Equal(t, "", actualHostID)
}

func Test_GpcHostIdIfZoneIdIsMissing(t *testing.T) {
	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.account.id", "project-id")
	attributes.PutStr("host.id", "instance-id")

	actualHostID := getGcpHostID(attributes)
	require.Equal(t, "", actualHostID)
}

func Test_GpcHostIdIfInstanceIdIsMissing(t *testing.T) {
	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.account.id", "project-id")
	attributes.PutStr("cloud.availability_zone", "zone-id")

	actualHostID := getGcpHostID(attributes)
	require.Equal(t, "", actualHostID)
}
