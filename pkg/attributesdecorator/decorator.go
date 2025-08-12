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

package attributesdecorator

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
)

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

func DecorateResourceAttributesByPluginIdentifiers[T Resource](collection ResourceCollection[T], properties *PluginProperties) {
	for i := 0; i < collection.Len(); i++ {
		resource := collection.At(i).Resource()
		resourceAttributes := resource.Attributes()
		properties.addPluginAttributes(resourceAttributes)
	}
}
