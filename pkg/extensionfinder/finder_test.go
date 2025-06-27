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
	l := zap.NewNop()
	host := &mockHost{
		extensions: make(map[component.ID]component.Component),
	}
	host.InsertExtension(t, "incorrectextension", &incorrectExtension{Name: "nomatch"})
	host.InsertExtension(t, "testingextension", &testingExtension{Name: "nomatch"})
	host.InsertExtension(t, extName, &testingExtension{Name: "match"})

	foundExt, err := FindExtension[*testingExtension](l, extName, host)

	assert.NoError(t, err, "should not return an error when finding the extension")
	assert.Equal(t, "match", foundExt.Name, "should return the correct extension")
}

func Test_FinsExtension_FailsOnCorrectTypeButWrongName(t *testing.T) {
	l := zap.NewNop()
	host := &mockHost{
		extensions: make(map[component.ID]component.Component),
	}
	host.InsertExtension(t, "testingextension", &testingExtension{Name: "nomatch"})
	host.InsertExtension(t, "testingextension/nonmatching", &testingExtension{Name: "nomatch"})

	_, err := FindExtension[*testingExtension](l, "testingextension/aliased", host)

	assert.ErrorContains(
		t,
		err,
		"extension \"testingextension/aliased\" not found",
		"should return an error when the extension is not found",
	)
}

func Test_FindExtension_FailsOnIncorrectType(t *testing.T) {
	l := zap.NewNop()
	host := &mockHost{
		extensions: make(map[component.ID]component.Component),
	}
	host.InsertExtension(t, "testingextension", &incorrectExtension{Name: "nomatch"})

	_, err := FindExtension[*testingExtension](l, "testingextension", host)

	assert.ErrorContains(
		t,
		err,
		"extension \"testingextension\" is not a *extensionfinder.testingExtension",
		"should return an error when the extension is not of the correct type",
	)
}
