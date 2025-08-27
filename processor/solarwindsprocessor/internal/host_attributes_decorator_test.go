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
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
)

func TestApplyAttributes_HostId_NonCloudHostWithOverrideHostId(t *testing.T) {
	pp := HostAttributesDecorator{
		IsRunInContainerd: false,
		ContainerID:       "",
		OnPremOverrideId:  "test-override-host-id",
		logger:            zap.NewNop(),
	}

	attributes := pcommon.NewMap()

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-override-host-id", hostId.Str())
}

func TestApplyAttributes_HostId_NonCloudHostWithoutoverrideHostId(t *testing.T) {
	pp := HostAttributesDecorator{
		IsRunInContainerd: false,
		ContainerID:       "",
		OnPremOverrideId:  "",
		logger:            zap.NewNop(),
	}

	attributes := pcommon.NewMap()
	attributes.PutStr(hostIDAttribute, "telemetry-host-id")

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "telemetry-host-id", hostId.Str()) // Should be empty string from OnPremOverrideID, not original
}

func TestApplyAttributes_HostId_CloudHostWithContainer(t *testing.T) {
	pp := HostAttributesDecorator{
		IsRunInContainerd: false,
		ContainerID:       "container-123",
		OnPremOverrideId:  "test-override-host-id",
		logger:            zap.NewNop(),
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "test-override-host-id", hostId.Str())
}

func TestApplyAttributes_HostId_CloudHostWithoutContainer(t *testing.T) {
	// Cloud host without container - hostId should remain unchanged
	pp := HostAttributesDecorator{
		IsRunInContainerd: false,
		ContainerID:       "", // No container
		OnPremOverrideId:  "test-override-host-id",
		logger:            zap.NewNop(),
	}

	attributes := pcommon.NewMap()
	attributes.PutStr("cloud.provider", "aws")
	attributes.PutStr("host.id", "original-host-id")

	pp.ApplyAttributes(attributes)

	hostId, exists := attributes.Get("host.id")
	require.True(t, exists)
	require.Equal(t, "original-host-id", hostId.Str())
}

func TestApplyAttributes_HostId_BiosUuidVsOverrideHostIdScenarios(t *testing.T) {
	testCases := map[string]struct {
		biosUuid         string
		biosUuidExists   bool
		onPremOverrideID string
		cloudProvider    string
		containerID      string
		expectedHostId   string
		description      string
	}{
		"bios uuid from telemetry exists and not empty - non cloud": {
			biosUuid:         "12345678-1234-1234-1234-123456789abc",
			biosUuidExists:   true,
			onPremOverrideID: "override-id-from-config",
			cloudProvider:    "",
			containerID:      "",
			expectedHostId:   "override-id-from-config",
			description:      "Should use override host ID from config when both BIOS UUID and override host ID are available on non-cloud host",
		},
		"bios uuid from telemetry exists but empty - non cloud": {
			biosUuid:         "",
			biosUuidExists:   true,
			onPremOverrideID: "override-id-from-config",
			cloudProvider:    "",
			containerID:      "",
			expectedHostId:   "override-id-from-config",
			description:      "Should use override host ID from config when BIOS UUID from telemetry is empty on non-cloud host",
		},
		"bios uuid from telemetry does not exist - non cloud": {
			biosUuid:         "",
			biosUuidExists:   false,
			onPremOverrideID: "override-id-from-config",
			cloudProvider:    "",
			containerID:      "",
			expectedHostId:   "override-id-from-config",
			description:      "Should use override host ID from config when BIOS UUID doesn't exist in telemetry on non-cloud host",
		},
		"bios uuid from telemetry exists - cloud host without container": {
			biosUuid:         "12345678-1234-1234-1234-123456789abc",
			biosUuidExists:   true,
			onPremOverrideID: "override-id-from-config",
			cloudProvider:    "aws",
			containerID:      "",
			expectedHostId:   "original-cloud-host-id",
			description:      "Should preserve original host ID on cloud host without container, ignoring both BIOS UUID and override host ID",
		},
		"bios uuid from telemetry exists - cloud host with container": {
			biosUuid:         "12345678-1234-1234-1234-123456789abc",
			biosUuidExists:   true,
			onPremOverrideID: "override-id-from-config",
			cloudProvider:    "aws",
			containerID:      "container-123",
			expectedHostId:   "override-id-from-config",
			description:      "Should use override host ID from config on cloud host with container (override host ID has precedence over BIOS UUID)",
		},
		"no bios uuid from telemetry - cloud host with container": {
			biosUuid:         "",
			biosUuidExists:   false,
			onPremOverrideID: "override-id-from-config",
			cloudProvider:    "aws",
			containerID:      "container-123",
			expectedHostId:   "override-id-from-config",
			description:      "Should use override host ID from config on cloud host with container when no BIOS UUID in telemetry",
		},
		"no override host id but bios uuid exists - non cloud": {
			biosUuid:         "12345678-1234-1234-1234-123456789abc",
			biosUuidExists:   true,
			onPremOverrideID: "",
			cloudProvider:    "",
			containerID:      "",
			expectedHostId:   "12345678-1234-1234-1234-123456789abc",
			description:      "Should use BIOS UUID from telemetry when override host ID is not configured on non-cloud host",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			pp := HostAttributesDecorator{
				IsRunInContainerd: false,
				ContainerID:       tc.containerID,
				OnPremOverrideId:  tc.onPremOverrideID,
				logger:            zap.NewNop(),
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

			pp := HostAttributesDecorator{
				IsRunInContainerd: tc.isRunInContainerd,
				ContainerID:       tc.containerID,
				OnPremOverrideId:  "test-client-id",
				logger:            zap.NewNop(),
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
	pp := HostAttributesDecorator{
		IsRunInContainerd: false,
		ContainerID:       "",
		OnPremOverrideId:  "test-override-host-id",
		logger:            zap.NewNop(),
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

func TestApplyAttributes_HostId_GcpScenarios(t *testing.T) {
	testCases := map[string]struct {
		cloudAccountID        string
		cloudAvailabilityZone string
		instanceID            string
		onPremOverrideID      string
		expectedHostId        string
		shouldHaveHostId      bool
		description           string
	}{
		"gcp with all required attributes": {
			cloudAccountID:        "my-project-123",
			cloudAvailabilityZone: "us-central1-a",
			instanceID:            "instance-456",
			onPremOverrideID:      "override-id",
			expectedHostId:        "my-project-123:us-central1-a:instance-456",
			shouldHaveHostId:      true,
			description:           "Should create GCP host ID from project:zone:instance when all attributes are present",
		},
		"gcp missing project id": {
			cloudAccountID:        "",
			cloudAvailabilityZone: "us-central1-a",
			instanceID:            "instance-456",
			onPremOverrideID:      "override-id",
			expectedHostId:        "",
			shouldHaveHostId:      false,
			description:           "Should remove host ID when project ID is missing",
		},
		"gcp missing availability zone": {
			cloudAccountID:        "my-project-123",
			cloudAvailabilityZone: "",
			instanceID:            "instance-456",
			onPremOverrideID:      "override-id",
			expectedHostId:        "",
			shouldHaveHostId:      false,
			description:           "Should remove host ID when availability zone is missing",
		},
		"gcp missing instance id": {
			cloudAccountID:        "my-project-123",
			cloudAvailabilityZone: "us-central1-a",
			instanceID:            "",
			onPremOverrideID:      "override-id",
			expectedHostId:        "",
			shouldHaveHostId:      false,
			description:           "Should remove host ID when instance ID is missing",
		},
		"gcp with all attributes missing": {
			cloudAccountID:        "",
			cloudAvailabilityZone: "",
			instanceID:            "",
			onPremOverrideID:      "override-id",
			expectedHostId:        "",
			shouldHaveHostId:      false,
			description:           "Should remove host ID when all GCP attributes are missing",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			pp := HostAttributesDecorator{
				IsRunInContainerd: false,
				ContainerID:       "",
				OnPremOverrideId:  tc.onPremOverrideID,
				logger:            zap.NewNop(),
			}

			attributes := pcommon.NewMap()
			attributes.PutStr("cloud.provider", "gcp")

			if tc.cloudAccountID != "" {
				attributes.PutStr("cloud.account.id", tc.cloudAccountID)
			}
			if tc.cloudAvailabilityZone != "" {
				attributes.PutStr("cloud.availability_zone", tc.cloudAvailabilityZone)
			}
			if tc.instanceID != "" {
				attributes.PutStr("host.id", tc.instanceID)
			}

			pp.ApplyAttributes(attributes)

			hostId, exists := attributes.Get("host.id")
			if tc.shouldHaveHostId {
				require.True(t, exists, tc.description)
				require.Equal(t, tc.expectedHostId, hostId.Str(), tc.description)
			} else {
				require.False(t, exists, tc.description)
			}
		})
	}
}
