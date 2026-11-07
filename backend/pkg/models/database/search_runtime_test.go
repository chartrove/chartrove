package database

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// Sharing a runtime without exclusive ownership or replacing its resource would
// mix search fields between simultaneous imports and between successive records.
func TestSearchExtractionIsolatesResources(t *testing.T) {
	for i := 0; i < 8; i++ {
		i := i
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			for j := 0; j < 3; j++ {
				name := fmt.Sprintf("Synthetic-%d-%d", i, j)
				patient := &FhirPatient{}
				raw := json.RawMessage(fmt.Sprintf(`{"resourceType":"Patient","id":"%d-%d","name":[{"family":%q}],"gender":"female"}`, i, j, name))
				require.NoError(t, patient.PopulateAndExtractSearchParameters(raw))
				var names []string
				require.NoError(t, json.Unmarshal(patient.Name, &names))
				require.Equal(t, []string{name}, names)

				empty := &FhirPatient{}
				require.NoError(t, empty.PopulateAndExtractSearchParameters(json.RawMessage(`{"resourceType":"Patient","id":"empty"}`)))
				require.Empty(t, empty.Name)
				require.Empty(t, empty.Gender)
			}
		})
	}
}

func TestSearchExtractionRejectsInvalidJSON(t *testing.T) {
	patient := &FhirPatient{}
	require.Error(t, patient.PopulateAndExtractSearchParameters(json.RawMessage(`{"resourceType":`)))
	require.NoError(t, patient.PopulateAndExtractSearchParameters(json.RawMessage(`{"resourceType":"Patient","gender":"male"}`)))
	var gender SearchParameterTokenType
	require.NoError(t, json.Unmarshal(patient.Gender, &gender))
	require.Equal(t, SearchParameterTokenType{{Code: "male"}}, gender)
}

// A runtime refresh must preserve supplied Unicode names without normalization
// or leaking the preceding resource through the shared program/runtime pool.
func TestSearchExtractionPreservesUnicodeNames(t *testing.T) {
	for _, family := range []string{"Synthetic Café", "Synthetic Cafe\u0301", "Synthetic 零", "Synthetic صفر"} {
		t.Run(family, func(t *testing.T) {
			raw, err := json.Marshal(map[string]interface{}{
				"resourceType": "Patient", "id": "synthetic-unicode-name",
				"name": []map[string]string{{"family": family}},
			})
			require.NoError(t, err)
			patient := &FhirPatient{}
			require.NoError(t, patient.PopulateAndExtractSearchParameters(raw))
			var names []string
			require.NoError(t, json.Unmarshal(patient.Name, &names))
			require.Equal(t, []string{family}, names)
		})
	}
}
