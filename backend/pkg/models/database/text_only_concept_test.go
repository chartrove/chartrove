package database

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestTokenExtractionPreservesTextOnlyConcepts(t *testing.T) {
	for _, test := range []struct{ name, concept, expected string }{
		{"text-only", `{"text":"Synthetic description Ω"}`, `[{"text":"Synthetic description Ω"}]`},
		{"empty-codings", `{"coding":[],"text":"Synthetic description Ω"}`, `[{"text":"Synthetic description Ω"}]`},
		{"null-codings", `{"coding":null,"text":"Synthetic description Ω"}`, `[{"text":"Synthetic description Ω"}]`},
		{"literal-markup", `{"text":"<b>literal</b> \"quoted\" Ω"}`, `[{"text":"<b>literal</b> \"quoted\" Ω"}]`},
		{"coding-display", `{"coding":[{"system":"urn:synthetic","code":"test","display":"Coding display"}],"text":"Concept text"}`, `[{"system":"urn:synthetic","code":"test","text":"Coding display"}]`},
		{"coding-fallback", `{"coding":[{"system":"urn:synthetic","code":"test"}],"text":"Concept text"}`, `[{"system":"urn:synthetic","code":"test","text":"Concept text"}]`},
		{"multiple-codings", `{"coding":[{"code":"first","display":"First"},{"code":"second"}],"text":"Fallback"}`, `[{"code":"first","text":"First"},{"code":"second","text":"Fallback"}]`},
		{"identifier", `{"system":"urn:synthetic","value":"id-0","type":{"text":"Synthetic ID"}}`, `[{"system":"urn:synthetic","code":"id-0","text":"Synthetic ID"}]`},
		{"primitive-false", `false`, `[{"code":false}]`},
		{"empty-concept", `{}`, "undefined"},
		{"empty-text", `{"coding":[],"text":""}`, "undefined"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var concept interface{}
			require.NoError(t, json.Unmarshal([]byte(test.concept), &concept))
			resource := map[string]interface{}{"resourceType": "Observation", "code": concept}
			vm, release, err := acquireSearchRuntime(resource)
			require.NoError(t, err)
			defer release()
			result, err := vm.RunString("extractTokenSearchParameters(fhirResource, 'Observation.code')")
			require.NoError(t, err)
			if test.expected == "undefined" {
				require.Equal(t, "undefined", result.String())
			} else {
				require.JSONEq(t, test.expected, result.String())
			}
		})
	}
}

func TestTextOnlyConceptsReachGeneratedModels(t *testing.T) {
	for _, kind := range []string{"AllergyIntolerance", "Condition", "Observation", "Medication"} {
		t.Run(kind, func(t *testing.T) {
			raw := []byte(fmt.Sprintf(`{"resourceType":%q,"id":"synthetic-only","code":{"text":"Synthetic concept Ω"},"reaction":[{"manifestation":[{"text":"Synthetic manifestation Ω"}]}],"unknownExtension":{"number":0,"flag":false}}`, kind))
			model, err := NewFhirResourceModelByType(kind)
			require.NoError(t, err)
			model.SetResourceRaw(datatypes.JSON(raw))
			require.NoError(t, model.PopulateAndExtractSearchParameters(raw))
			encoded, err := json.Marshal(model)
			require.NoError(t, err)
			var columns map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(encoded, &columns))
			require.JSONEq(t, `[{"text":"Synthetic concept Ω"}]`, string(columns["code"]))
			if kind == "AllergyIntolerance" {
				require.JSONEq(t, `[{"text":"Synthetic manifestation Ω"}]`, string(columns["manifestation"]))
			}
			require.Equal(t, raw, []byte(model.GetResourceRaw()), "derived columns must not rewrite raw records")
		})
	}
}
