package solarwindsprocessor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

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
