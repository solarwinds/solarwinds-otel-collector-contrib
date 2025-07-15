package internal

import (
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"testing"
)

func TestUpdate_Relationship(t *testing.T) {
	logRecords := plog.NewLogRecordSlice()

	srcIds := pcommon.NewMap()
	srcIds.PutStr("cluster.uid", "cluster123")

	destIds := pcommon.NewMap()
	destIds.PutStr("namespace.uid", "namespace456")

	attrs := pcommon.NewMap()
	attrs.PutStr("relationship.attr", "test-value")

	relationship := Relationship{
		Action: "update",
		Type:   "Has",
		Source: Entity{
			Type: "KubernetesCluster",
			IDs:  srcIds,
		},
		Destination: Entity{
			Type: "KubernetesNamespace",
			IDs:  destIds,
		},
		Attributes: attrs,
	}

	relationship.Update(&logRecords)

	logRecord := logRecords.At(0)
	assert.Equal(t, 7, logRecord.Attributes().Len())

	actualEntityEventType, _ := logRecord.Attributes().Get(entityEventType)
	assert.Equal(t, relationshipUpdateEventType, actualEntityEventType.Str())

	actualRelationshipType, _ := logRecord.Attributes().Get(relationshipType)
	assert.Equal(t, "Has", actualRelationshipType.Str())

	actualSrcEntityType, _ := logRecord.Attributes().Get(srcEntityType)
	assert.Equal(t, "KubernetesCluster", actualSrcEntityType.Str())

	actualDestEntityType, _ := logRecord.Attributes().Get(destEntityType)
	assert.Equal(t, "KubernetesNamespace", actualDestEntityType.Str())

	actualSrcIds, _ := logRecord.Attributes().Get(relationshipSrcEntityIds)
	assert.Equal(t, 1, actualSrcIds.Map().Len())
	actualClusterUID, _ := actualSrcIds.Map().Get("cluster.uid")
	assert.Equal(t, "cluster123", actualClusterUID.Str())

	actualDestIds, _ := logRecord.Attributes().Get(relationshipDestEntityIds)
	assert.Equal(t, 1, actualDestIds.Map().Len())
	actualNamespaceUID, _ := actualDestIds.Map().Get("namespace.uid")
	assert.Equal(t, "namespace456", actualNamespaceUID.Str())

	actualAttributes, _ := logRecord.Attributes().Get(relationshipAttributes)
	assert.Equal(t, 1, actualAttributes.Map().Len())
	actualRelationshipAttr, _ := actualAttributes.Map().Get("relationship.attr")
	assert.Equal(t, "test-value", actualRelationshipAttr.Str())
}

func TestDelete_Relationship(t *testing.T) {
	logRecords := plog.NewLogRecordSlice()

	srcIds := pcommon.NewMap()
	srcIds.PutStr("cluster.uid", "cluster123")

	destIds := pcommon.NewMap()
	destIds.PutStr("namespace.uid", "namespace456")

	relationship := Relationship{
		Action: "delete",
		Type:   "Has",
		Source: Entity{
			Type: "KubernetesCluster",
			IDs:  srcIds,
		},
		Destination: Entity{
			Type: "KubernetesNamespace",
			IDs:  destIds,
		},
		Attributes: pcommon.NewMap(), // Empty attributes for delete
	}

	relationship.Delete(&logRecords)

	logRecord := logRecords.At(0)
	assert.Equal(t, 6, logRecord.Attributes().Len())

	actualEntityEventType, _ := logRecord.Attributes().Get(entityEventType)
	assert.Equal(t, relationshipDeleteEventType, actualEntityEventType.Str())

	actualRelationshipType, _ := logRecord.Attributes().Get(relationshipType)
	assert.Equal(t, "Has", actualRelationshipType.Str())

	actualSrcEntityType, _ := logRecord.Attributes().Get(srcEntityType)
	assert.Equal(t, "KubernetesCluster", actualSrcEntityType.Str())

	actualDestEntityType, _ := logRecord.Attributes().Get(destEntityType)
	assert.Equal(t, "KubernetesNamespace", actualDestEntityType.Str())

	actualSrcIds, _ := logRecord.Attributes().Get(relationshipSrcEntityIds)
	assert.Equal(t, 1, actualSrcIds.Map().Len())
	actualClusterUID, _ := actualSrcIds.Map().Get("cluster.uid")
	assert.Equal(t, "cluster123", actualClusterUID.Str())

	actualDestIds, _ := logRecord.Attributes().Get(relationshipDestEntityIds)
	assert.Equal(t, 1, actualDestIds.Map().Len())
	actualNamespaceUID, _ := actualDestIds.Map().Get("namespace.uid")
	assert.Equal(t, "namespace456", actualNamespaceUID.Str())

	// Verify relationship attributes are not present in delete
	_, exists := logRecord.Attributes().Get(relationshipAttributes)
	assert.False(t, exists)
}
