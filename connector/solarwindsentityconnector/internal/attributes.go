package internal

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"strings"
)

type Attributes struct {
	Source      map[string]pcommon.Value
	Destination map[string]pcommon.Value
	Common      map[string]pcommon.Value
}

func IdentifyAttributes(resourceAttrs pcommon.Map, srcPrefix, destPrefix string) Attributes {
	attrs := Attributes{
		Source:      make(map[string]pcommon.Value),
		Destination: make(map[string]pcommon.Value),
		Common:      make(map[string]pcommon.Value),
	}
	for k, v := range resourceAttrs.All() {
		switch {
		case srcPrefix != "" && strings.HasPrefix(k, srcPrefix):
			attrs.Source[getWithoutPrefix(srcPrefix, k)] = v
		case destPrefix != "" && strings.HasPrefix(k, destPrefix):
			attrs.Destination[getWithoutPrefix(destPrefix, k)] = v
		default:
			attrs.Common[k] = v
		}
	}

	return attrs
}

func getWithoutPrefix(prefix, key string) string {
	return strings.TrimPrefix(key, prefix)
}
