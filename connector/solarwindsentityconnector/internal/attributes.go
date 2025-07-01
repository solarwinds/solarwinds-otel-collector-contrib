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
		case strings.HasPrefix(k, srcPrefix):
			attrs.Source[k] = v
		case strings.HasPrefix(k, destPrefix):
			attrs.Destination[k] = v
		default:
			attrs.Common[k] = v
		}
	}

	return attrs
}
