package swag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sv-tools/openapi/spec"
)

func TestFingerprintV3(t *testing.T) {
	t.Parallel()

	a := &spec.Parameter{Name: "id", In: "path", Required: true}
	b := &spec.Parameter{Name: "id", In: "path", Required: true}
	c := &spec.Parameter{Name: "id", In: "query", Required: true}

	fpA, err := fingerprintV3(a)
	require.NoError(t, err)

	fpB, err := fingerprintV3(b)
	require.NoError(t, err)

	fpC, err := fingerprintV3(c)
	require.NoError(t, err)

	assert.Equal(t, fpA, fpB, "structurally identical objects must fingerprint the same")
	assert.NotEqual(t, fpA, fpC, "objects differing in even one field must fingerprint differently")
}

func TestUniqueComponentNameV3(t *testing.T) {
	t.Parallel()

	existing := map[string]bool{}

	first := uniqueComponentNameV3("Pet", existing)
	assert.Equal(t, "Pet", first)

	second := uniqueComponentNameV3("Pet", existing)
	assert.Equal(t, "Pet2", second, "collision must be resolved with a numeric suffix")

	third := uniqueComponentNameV3("Pet", existing)
	assert.Equal(t, "Pet3", third)

	assert.True(t, existing["Pet"])
	assert.True(t, existing["Pet2"])
	assert.True(t, existing["Pet3"])
}

func TestUniqueComponentNameV3EmptyPreferred(t *testing.T) {
	t.Parallel()

	existing := map[string]bool{}
	name := uniqueComponentNameV3("", existing)
	assert.Equal(t, "Component", name)
}

func TestExportedNameV3(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "XRequestId", exportedNameV3("X-Request-Id"))
	assert.Equal(t, "Id", exportedNameV3("id"))
	assert.Equal(t, "", exportedNameV3(""))
	assert.Equal(t, "Abc123", exportedNameV3("abc123"))
}

func TestSchemaRefNameV3(t *testing.T) {
	t.Parallel()

	ref := spec.NewSchemaRef(spec.NewRef("#/components/schemas/model.Pet"))
	assert.Equal(t, "ModelPet", schemaRefNameV3(ref))

	inline := spec.NewSchemaSpec()
	assert.Equal(t, "", schemaRefNameV3(inline))

	assert.Equal(t, "", schemaRefNameV3(nil))
}
