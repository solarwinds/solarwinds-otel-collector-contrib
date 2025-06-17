package attributesdecorator

import "go.opentelemetry.io/collector/pdata/pcommon"

type Resource interface {
	Resource() pcommon.Resource
}

type ResourceCollection[T Resource] interface {
	At(index int) T
	Len() int
}

func DecorateResourceAttributes[T Resource](collection ResourceCollection[T], atts map[string]string) {
	collectionLen := collection.Len()
	if collectionLen == 0 {
		return
	}

	for i := 0; i < collectionLen; i++ {
		resource := collection.At(i).Resource()
		resourceAttributes := resource.Attributes()
		for key, value := range atts {
			resourceAttributes.PutStr(key, value)
		}
	}
}
