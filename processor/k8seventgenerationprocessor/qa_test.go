package k8seventgenerationprocessor

import (
	"testing"

	"github.com/solarwinds/solarwinds-otel-collector-contrib/processor/k8seventgenerationprocessor/internal/constants"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestEmptyTagScenario(t *testing.T) {
	// Helper to verify empty tag array
	verifyEmptyTagArray := func(t *testing.T, attrs pcommon.Map) {
		entityAttrs := getMapValue(t, attrs, "otel.entity.attributes")
		tagsSlice, tagsExists := entityAttrs.Get(constants.AttributeContainerImageTags)
		assert.True(t, tagsExists, "container.image.tags should exist")
		assert.Equal(t, pcommon.ValueTypeSlice, tagsSlice.Type(), "container.image.tags should be a slice")
		assert.Equal(t, 0, tagsSlice.Slice().Len(), "tags array should be empty")
	}

	t.Run("PodWithEmptyTag", func(t *testing.T) {
		// Pod manifest where image parsing fails (e.g. empty string or invalid format),
		// but ImageID is present in status. This results in empty Name and empty Tag.

		podManifestNoTag := `{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod-no-tag","namespace":"test-namespace","uid":"uid-no-tag"},"spec":{"containers":[{"image":"","name":"test-container"}]},"status":{"containerStatuses":[{"containerID":"test-container-id","image":"","imageID":"sha256:digest123","name":"test-container"}]}}`

		l := generateManifestLogs("Pod", podManifestNoTag)
		consumer, err := startAndConsumeLogs(t, l)
		assert.NoError(t, err)

		result := consumer.AllLogs()
		assert.Len(t, result, 1)

		// Find the container image log
		newLog := result[0].ResourceLogs().At(1)
		scopeLogs := newLog.ScopeLogs().At(0)

		var foundImage bool
		for _, lr := range scopeLogs.LogRecords().All() {
			attrs := lr.Attributes()
			eventType := getStringValue(t, attrs, "otel.entity.event.type")
			if eventType == "entity_state" {
				entityType := getStringValue(t, attrs, "otel.entity.type")
				if entityType == "KubernetesContainerImage" {
					foundImage = true
					verifyEmptyTagArray(t, attrs)
					// Also verify name is empty because parsing failed
					ids := getMapValue(t, attrs, "otel.entity.id")
					assert.Equal(t, "", getStringValue(t, ids, "container.image.name"))
				}
			}
		}
		assert.True(t, foundImage, "Should have found a container image entity")
	})

	t.Run("VulnerabilityReportWithEmptyTag", func(t *testing.T) {
		// VulnerabilityReport where artifact.tag is empty
		vulnReportNoTag := `{
			"apiVersion": "aquasecurity.github.io/v1alpha1",
			"kind": "VulnerabilityReport",
			"metadata": {
				"name": "test-report",
				"namespace": "test-namespace"
			},
			"report": {
				"artifact": {
					"digest": "sha256:digest456",
					"repository": "test-repo",
					"tag": ""
				},
				"scanner": {
					"name": "Trivy",
					"vendor": "Aqua Security",
					"version": "0.40.0"
				},
				"vulnerabilities": []
			}
		}`

		l := generateManifestLogs("VulnerabilityReport", vulnReportNoTag)
		consumer, err := startAndConsumeLogs(t, l)
		assert.NoError(t, err)

		result := consumer.AllLogs()
		assert.Len(t, result, 1)

		// Find the container image log
		newLog := result[0].ResourceLogs().At(1)
		scopeLogs := newLog.ScopeLogs().At(0)

		var foundImage bool
		for _, lr := range scopeLogs.LogRecords().All() {
			attrs := lr.Attributes()
			eventType := getStringValue(t, attrs, "otel.entity.event.type")
			if eventType == "entity_state" {
				entityType := getStringValue(t, attrs, "otel.entity.type")
				if entityType == "KubernetesContainerImage" {
					foundImage = true
					verifyEmptyTagArray(t, attrs)
				}
			}
		}
		assert.True(t, foundImage, "Should have found a container image entity")
	})
}
