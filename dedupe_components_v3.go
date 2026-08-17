package swag

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/sv-tools/openapi/spec"
)

// deduplicateComponentsV3 extracts parameters, request bodies, responses, and headers that
// are structurally identical across two or more operations into shared components
// (components.parameters/requestBodies/responses/headers), replacing each occurrence with a
// $ref — the same "define once, reference everywhere" outcome that named Go types already
// get for free via components.schemas, just keyed by a content fingerprint instead of a Go
// type identity, since an object built from a @Param/@Success line has no such identity to
// key on.
//
// Runs once, after every operation (including webhooks and callbacks) is fully built, so it
// sees the whole document before deciding what's actually repeated. Header dedup must run
// before response dedup: response fingerprints include their Headers map, so headers need to
// already be in their final ($ref or inline) shape before a response's fingerprint is taken.
func (p *Parser) deduplicateComponentsV3() error {
	operations := collectAllOperationsV3(p)

	dedupeHeadersV3(p, operations)
	dedupeResponsesV3(p, operations)
	dedupeRequestBodiesV3(p, operations)
	dedupeParametersV3(p, operations)

	return nil
}

// collectAllOperationsV3 gathers every operation in the document: paths, webhooks, and
// operations nested one level under any operation's callbacks.
func collectAllOperationsV3(p *Parser) []*spec.Operation {
	var operations []*spec.Operation

	collectFromPathItem := func(item *spec.PathItem) {
		for method := range allMethod {
			if op := readRouteMethodOpV3(item, method); op != nil {
				operations = append(operations, op)
			}
		}
	}

	if p.openAPI.Paths != nil {
		for _, pathItem := range p.openAPI.Paths.Spec.Paths {
			collectFromPathItem(pathItem.Spec.Spec)
		}
	}

	for _, pathItem := range p.openAPI.WebHooks {
		collectFromPathItem(pathItem.Spec.Spec)
	}

	// Callback operations are collected in a second step so we don't mutate the slice
	// we're ranging over.
	var callbackOps []*spec.Operation

	for _, op := range operations {
		for _, callback := range op.Callbacks {
			if callback == nil || callback.Spec == nil {
				continue
			}

			for _, pathItem := range callback.Spec.Spec.Callback {
				if pathItem == nil || pathItem.Spec == nil {
					continue
				}

				for method := range allMethod {
					if cbOp := readRouteMethodOpV3(pathItem.Spec.Spec, method); cbOp != nil {
						callbackOps = append(callbackOps, cbOp)
					}
				}
			}
		}
	}

	return append(operations, callbackOps...)
}

// fingerprintV3 returns a stable content hash for any JSON-marshalable value, used to detect
// structurally identical parameters/request bodies/responses/headers regardless of which
// operation they came from. Go's encoding/json sorts map keys when marshaling, so this is
// deterministic across calls.
func fingerprintV3(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:]), nil
}

// uniqueComponentNameV3 returns preferred if it's free in existing, else preferred2,
// preferred3, ... until one is, and reserves whichever name it returns.
func uniqueComponentNameV3(preferred string, existing map[string]bool) string {
	if preferred == "" {
		preferred = "Component"
	}

	name := preferred
	for i := 2; existing[name]; i++ {
		name = fmt.Sprintf("%s%d", preferred, i)
	}

	existing[name] = true

	return name
}

// exportedNameV3 upper-cases the first rune, and strips characters that aren't safe in a
// component name (component names appear directly in a URI fragment).
func exportedNameV3(s string) string {
	var b strings.Builder

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}

	out := b.String()
	if out == "" {
		return ""
	}

	r := []rune(out)
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}

// schemaRefNameV3 returns the component name a schema $refs to (e.g. "model.Pet" out of
// "#/components/schemas/model.Pet"), or "" if schema is nil or isn't a $ref.
func schemaRefNameV3(schema *spec.RefOrSpec[spec.Schema]) string {
	if schema == nil || schema.Ref == nil {
		return ""
	}

	ref := schema.Ref.Ref
	if pos := strings.LastIndexByte(ref, '/'); pos >= 0 {
		return exportedNameV3(ref[pos+1:])
	}

	return ""
}

// firstMediaTypeSchemaV3 returns the schema of the first media type in content, preferring
// application/json when present, for naming purposes only.
func firstMediaTypeSchemaV3(content map[string]*spec.Extendable[spec.MediaType]) *spec.RefOrSpec[spec.Schema] {
	if media, ok := content[mimeTypeJSON]; ok && media != nil && media.Spec != nil {
		return media.Spec.Schema
	}

	for _, media := range content {
		if media != nil && media.Spec != nil && media.Spec.Schema != nil {
			return media.Spec.Schema
		}
	}

	return nil
}

func dedupeParametersV3(p *Parser, operations []*spec.Operation) {
	type occurrence struct {
		op  *spec.Operation
		idx int
	}

	groups := map[string][]occurrence{}
	firstParam := map[string]*spec.RefOrSpec[spec.Extendable[spec.Parameter]]{}

	for _, op := range operations {
		for i, param := range op.Parameters {
			if param == nil || param.Ref != nil || param.Spec == nil || param.Spec.Spec == nil {
				continue
			}

			fp, err := fingerprintV3(param.Spec.Spec)
			if err != nil {
				continue
			}

			groups[fp] = append(groups[fp], occurrence{op: op, idx: i})
			if _, ok := firstParam[fp]; !ok {
				firstParam[fp] = param
			}
		}
	}

	existingNames := map[string]bool{}
	for name := range p.openAPI.Components.Spec.Parameters {
		existingNames[name] = true
	}

	for fp, occs := range groups {
		if len(occs) < 2 {
			continue
		}

		param := firstParam[fp]
		preferred := exportedNameV3(param.Spec.Spec.Name) + exportedNameV3(param.Spec.Spec.In)
		name := uniqueComponentNameV3(preferred, existingNames)

		if p.openAPI.Components.Spec.Parameters == nil {
			p.openAPI.Components.Spec.Parameters = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.Parameter]])
		}

		p.openAPI.Components.Spec.Parameters[name] = param

		ref := spec.NewParameterRef(spec.NewRef("#/components/parameters/" + name))
		for _, occ := range occs {
			occ.op.Parameters[occ.idx] = ref
		}
	}
}

func dedupeRequestBodiesV3(p *Parser, operations []*spec.Operation) {
	groups := map[string][]*spec.Operation{}
	first := map[string]*spec.RefOrSpec[spec.Extendable[spec.RequestBody]]{}

	for _, op := range operations {
		rb := op.RequestBody
		if rb == nil || rb.Ref != nil || rb.Spec == nil || rb.Spec.Spec == nil {
			continue
		}

		fp, err := fingerprintV3(rb.Spec.Spec)
		if err != nil {
			continue
		}

		groups[fp] = append(groups[fp], op)
		if _, ok := first[fp]; !ok {
			first[fp] = rb
		}
	}

	existingNames := map[string]bool{}
	for name := range p.openAPI.Components.Spec.RequestBodies {
		existingNames[name] = true
	}

	for fp, ops := range groups {
		if len(ops) < 2 {
			continue
		}

		rb := first[fp]
		preferred := schemaRefNameV3(firstMediaTypeSchemaV3(rb.Spec.Spec.Content))
		name := uniqueComponentNameV3(preferred+"RequestBody", existingNames)

		if p.openAPI.Components.Spec.RequestBodies == nil {
			p.openAPI.Components.Spec.RequestBodies = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.RequestBody]])
		}

		p.openAPI.Components.Spec.RequestBodies[name] = rb

		ref := spec.NewRequestBodyRef(spec.NewRef("#/components/requestBodies/" + name))
		for _, op := range ops {
			op.RequestBody = ref
		}
	}
}

func dedupeResponsesV3(p *Parser, operations []*spec.Operation) {
	type occurrence struct {
		responses *spec.Responses
		code      string // "" means Default
	}

	groups := map[string][]occurrence{}
	first := map[string]*spec.RefOrSpec[spec.Extendable[spec.Response]]{}

	for _, op := range operations {
		if op.Responses == nil || op.Responses.Spec == nil {
			continue
		}

		responses := op.Responses.Spec

		visit := func(code string, resp *spec.RefOrSpec[spec.Extendable[spec.Response]]) {
			if resp == nil || resp.Ref != nil || resp.Spec == nil || resp.Spec.Spec == nil {
				return
			}

			fp, err := fingerprintV3(resp.Spec.Spec)
			if err != nil {
				return
			}

			groups[fp] = append(groups[fp], occurrence{responses: responses, code: code})
			if _, ok := first[fp]; !ok {
				first[fp] = resp
			}
		}

		visit("", responses.Default)
		for code, resp := range responses.Response {
			visit(code, resp)
		}
	}

	existingNames := map[string]bool{}
	for name := range p.openAPI.Components.Spec.Responses {
		existingNames[name] = true
	}

	for fp, occs := range groups {
		if len(occs) < 2 {
			continue
		}

		resp := first[fp]
		preferred := schemaRefNameV3(firstMediaTypeSchemaV3(resp.Spec.Spec.Content))
		name := uniqueComponentNameV3(preferred+"Response", existingNames)

		if p.openAPI.Components.Spec.Responses == nil {
			p.openAPI.Components.Spec.Responses = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.Response]])
		}

		p.openAPI.Components.Spec.Responses[name] = resp

		ref := spec.NewResponseRef(spec.NewRef("#/components/responses/" + name))
		for _, occ := range occs {
			if occ.code == "" {
				occ.responses.Default = ref
			} else {
				occ.responses.Response[occ.code] = ref
			}
		}
	}
}

func dedupeHeadersV3(p *Parser, operations []*spec.Operation) {
	type occurrence struct {
		headers map[string]*spec.RefOrSpec[spec.Extendable[spec.Header]]
		key     string
	}

	groups := map[string][]occurrence{}
	first := map[string]*spec.RefOrSpec[spec.Extendable[spec.Header]]{}
	preferredName := map[string]string{}

	visitResponse := func(resp *spec.RefOrSpec[spec.Extendable[spec.Response]]) {
		if resp == nil || resp.Ref != nil || resp.Spec == nil || resp.Spec.Spec == nil {
			return
		}

		for key, header := range resp.Spec.Spec.Headers {
			if header == nil || header.Ref != nil || header.Spec == nil || header.Spec.Spec == nil {
				continue
			}

			fp, err := fingerprintV3(header.Spec.Spec)
			if err != nil {
				continue
			}

			groups[fp] = append(groups[fp], occurrence{headers: resp.Spec.Spec.Headers, key: key})
			if _, ok := first[fp]; !ok {
				first[fp] = header
				preferredName[fp] = key
			}
		}
	}

	for _, op := range operations {
		if op.Responses == nil || op.Responses.Spec == nil {
			continue
		}

		visitResponse(op.Responses.Spec.Default)
		for _, resp := range op.Responses.Spec.Response {
			visitResponse(resp)
		}
	}

	existingNames := map[string]bool{}
	for name := range p.openAPI.Components.Spec.Headers {
		existingNames[name] = true
	}

	for fp, occs := range groups {
		if len(occs) < 2 {
			continue
		}

		header := first[fp]
		name := uniqueComponentNameV3(exportedNameV3(preferredName[fp]), existingNames)

		if p.openAPI.Components.Spec.Headers == nil {
			p.openAPI.Components.Spec.Headers = make(map[string]*spec.RefOrSpec[spec.Extendable[spec.Header]])
		}

		p.openAPI.Components.Spec.Headers[name] = header

		ref := spec.NewHeaderRef(spec.NewRef("#/components/headers/" + name))
		for _, occ := range occs {
			occ.headers[occ.key] = ref
		}
	}
}
