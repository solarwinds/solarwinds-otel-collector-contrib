package extensionfinder

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type testingExtension struct {
	Name string
}

type incorrectExtension struct {
	Name string
}

func (*incorrectExtension) Shutdown(context.Context) error              { return nil }
func (*incorrectExtension) Start(context.Context, component.Host) error { return nil }

type mockHost struct {
	extensions map[component.ID]component.Component
}

func (*testingExtension) Shutdown(context.Context) error              { return nil }
func (*testingExtension) Start(context.Context, component.Host) error { return nil }

func (m *mockHost) GetExtensions() map[component.ID]component.Component {
	return m.extensions
}

func (m *mockHost) InsertExtension(
	t *testing.T,
	extensionName string,
	ext component.Component,
) {
	id := new(component.ID)
	if err := id.UnmarshalText([]byte(extensionName)); err != nil {
		assert.Failf(t, "failed to parse extension ID %q", extensionName)
	}
	m.extensions[*id] = ext
}

func Test_FindExtension_SuceedsOnExactNameAndType(t *testing.T) {
	extName := "testingextension/aliased"

	host := &mockHost{
		extensions: make(map[component.ID]component.Component),
	}
	host.InsertExtension(t, "incorrectextension", &incorrectExtension{Name: "nomatch"})
	host.InsertExtension(t, "testingextension", &testingExtension{Name: "nomatch"})
	host.InsertExtension(t, extName, &testingExtension{Name: "match"})

	foundExt, err := FindExtension[*testingExtension](new(zap.Logger), extName, host)
	assert.NoError(t, err, "should not return an error when finding the extension")

	assert.Equal(t, "match", foundExt.Name, "should return the correct extension")
}

/*func Test_FinsExtension_FailsOnCorrectTypeButWrongName(t *testing.T) {
	host := &mockHost{
		extensions: make(map[component.ID]component.Component),
	}
	host.InsertExtension(t, "incorrectextension", &incorrectExtension{Name: "nomatch"})
	host.InsertExtension(t, "testingextension", &testingExtension{Name: "nomatch"})
	host.InsertExtension(t, "testingextension/nonmatching", &testingExtension{Name: "nomatch"})

	_, err := FindExtension[*testingExtension](new(zap.Logger), "testingextension/aliased", host)
	assert.Error()
	assert.Error(t, err, "should not return an error when finding the extension")
}

func Test_FindExtension_FailsOnWrongName(t *testing.T) {
}
*/
