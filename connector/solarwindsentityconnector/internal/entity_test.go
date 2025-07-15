package internal

import (
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"testing"
)

func TestUpdate_Entity(t *testing.T) {
	logRecords := plog.NewLogRecordSlice()
	ids := pcommon.NewMap()
	ids.PutStr("cluster.uid", "cluster123")
	attrs := pcommon.NewMap()
	attrs.PutStr("cluster.name", "test-cluster")
	entity := Entity{
		Action:     "update", // not used but it would be present in the entity
		Type:       "KubernetesCluster",
		IDs:        ids,
		Attributes: attrs,
	}

	entity.Update(&logRecords)

	logRecord := logRecords.At(0)
	assert.Equal(t, 4, logRecord.Attributes().Len())

	actualEntityEventType, _ := logRecord.Attributes().Get(entityEventType)
	assert.Equal(t, entityUpdateEventType, actualEntityEventType.Str())

	actualRelationshipType, _ := logRecord.Attributes().Get(entityType)
	assert.Equal(t, "KubernetesCluster", actualRelationshipType.Str())

	actualIds, _ := logRecord.Attributes().Get(entityIds)
	assert.Equal(t, 1, actualIds.Map().Len())
	actualClusterUID, _ := actualIds.Map().Get("cluster.uid")
	assert.Equal(t, "cluster123", actualClusterUID.Str())

	actualAttributes, _ := logRecord.Attributes().Get(entityAttributes)
	assert.Equal(t, 1, actualAttributes.Map().Len())
	actualClusterName, _ := actualAttributes.Map().Get("cluster.name")
	assert.Equal(t, "test-cluster", actualClusterName.Str())
}

func TestDelete_Entity(t *testing.T) {
	logRecords := plog.NewLogRecordSlice()
	ids := pcommon.NewMap()
	ids.PutStr("cluster.uid", "cluster123")
	attrs := pcommon.NewMap()
	attrs.PutStr("cluster.name", "test-cluster")
	entity := Entity{
		Action:     "delete",
		Type:       "KubernetesCluster",
		IDs:        ids,
		Attributes: attrs,
	}

	entity.Delete(&logRecords)

	logRecord := logRecords.At(0)
	assert.Equal(t, 5, logRecord.Attributes().Len())

	actualEntityEventType, _ := logRecord.Attributes().Get(entityEventType)
	assert.Equal(t, entityDeleteEventType, actualEntityEventType.Str())

	actualRelationshipType, _ := logRecord.Attributes().Get(entityType)
	assert.Equal(t, "KubernetesCluster", actualRelationshipType.Str())

	actualIds, _ := logRecord.Attributes().Get(entityIds)
	assert.Equal(t, 1, actualIds.Map().Len())
	actualClusterUID, _ := actualIds.Map().Get("cluster.uid")
	assert.Equal(t, "cluster123", actualClusterUID.Str())

	actualAttributes, _ := logRecord.Attributes().Get(entityAttributes)
	assert.Equal(t, 1, actualAttributes.Map().Len())
	actualClusterName, _ := actualAttributes.Map().Get("cluster.name")
	assert.Equal(t, "test-cluster", actualClusterName.Str())
}
