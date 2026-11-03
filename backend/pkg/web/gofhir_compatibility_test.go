package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	sourceModels "github.com/fastenhealth/fasten-sources/clients/models"
	"github.com/fastenhealth/gofhir-models/fhir401"
	fhir401utils "github.com/fastenhealth/gofhir-models/fhir401/utils"
	"github.com/fastenhealth/gofhir-models/fhir430"
	fhir430utils "github.com/fastenhealth/gofhir-models/fhir430/utils"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Compile the actual provider SDK contract, including contained-resource types.
var _ sourceModels.ResourceInterface = fhir401.Patient{}
var _ sourceModels.ResourceInterface = fhir401.Observation{}
var _ sourceModels.ResourceInterface = fhir430.Patient{}
var _ sourceModels.ResourceInterface = fhir430.Observation{}

func strictFHIRObject(t *testing.T, raw []byte) any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func TestFHIRModelGenderContract(t *testing.T) {
	for _, version := range []string{"R4", "R4B"} {
		t.Run(version, func(t *testing.T) {
			parse := func(raw []byte) (any, error) {
				if version == "R4" {
					return fhir401utils.MapToResource(raw, false)
				}
				return fhir430utils.MapToResource(raw, false)
			}
			for input, want := range map[string]string{"male": "male", "female": "female", "other": "other", "unknown": "unknown"} {
				t.Run(input, func(t *testing.T) {
					raw := []byte(fmt.Sprintf(`{"resourceType":"Patient","id":"synthetic-model-patient","gender":%q}`, input))
					model, err := parse(raw)
					if err != nil {
						t.Fatal(err)
					}
					sdk, ok := model.(sourceModels.ResourceInterface)
					if !ok {
						t.Fatal("provider SDK contract is incompatible")
					}
					kind, id := sdk.ResourceRef()
					if kind != "Patient" || id == nil || *id != "synthetic-model-patient" {
						t.Fatal("resource identity changed")
					}
					encoded, err := json.Marshal(model)
					if err != nil {
						t.Fatal(err)
					}
					object := strictFHIRObject(t, encoded).(map[string]any)
					if object["gender"] != want {
						t.Fatalf("gender=%v, want %s", object["gender"], want)
					}
				})
			}
			for _, input := range []string{`"clinically-invented"`, `""`, `42`} {
				t.Run("reject-"+input, func(t *testing.T) {
					_, err := parse([]byte(`{"resourceType":"Patient","id":"synthetic-model-patient","gender":` + input + `}`))
					if err == nil {
						t.Fatal("invalid clinical gender accepted")
					}
				})
			}
		})
	}
}

func TestFHIRModelKnownFieldsRoundTrip(t *testing.T) {
	raw := []byte(`{"resourceType":"Patient","id":"synthetic-model-patient","name":[{"text":"Synthetic عربي"}],"active":false,"deceasedBoolean":false,"multipleBirthInteger":0,"gender":"unknown","birthDate":"2000-02","meta":{"lastUpdated":"2024-02-29T23:59:59.123456+03:00"},"extension":[{"url":"https://example.invalid/false","valueBoolean":false}]}`)
	for _, version := range []string{"R4", "R4B"} {
		t.Run(version, func(t *testing.T) {
			var model any
			var err error
			if version == "R4" {
				model, err = fhir401utils.MapToResource(raw, false)
			} else {
				model, err = fhir430utils.MapToResource(raw, false)
			}
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(model)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(strictFHIRObject(t, raw), strictFHIRObject(t, encoded)) {
				t.Fatal("known fields changed during typed FHIR round-trip")
			}
		})
	}
}

func TestFHIRModelEscapedGenderContract(t *testing.T) {
	// This is the standard code "male" encoded using a valid JSON escape,
	// rather than a nonstandard shorthand clinical value.
	raw := []byte(`{"resourceType":"Patient","id":"synthetic-model-patient","gender":"ma\u006ce"}`)
	for _, version := range []string{"R4", "R4B"} {
		t.Run(version, func(t *testing.T) {
			var model any
			var err error
			if version == "R4" {
				model, err = fhir401utils.MapToResource(raw, false)
			} else {
				model, err = fhir430utils.MapToResource(raw, false)
			}
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(model)
			if err != nil {
				t.Fatal(err)
			}
			object := strictFHIRObject(t, encoded).(map[string]any)
			if object["gender"] != "male" || object["resourceType"] != "Patient" {
				t.Fatal("valid escaped standard code or resource type changed")
			}
		})
	}
}

func TestEmbeddedFHIRModelImportRawReopen(t *testing.T) {
	blocked := &blockedTransport{}
	previous := http.DefaultTransport
	http.DefaultTransport = blocked
	t.Cleanup(func() { http.DefaultTransport = previous })
	for _, gender := range []string{"male", "female", "other", "unknown"} {
		t.Run(gender, func(t *testing.T) {
			t.Setenv("TMPDIR", t.TempDir())
			db := filepath.Join(t.TempDir(), "synthetic-encrypted.db")
			key := strings.Repeat("synthetic-fhir-key", 3)
			cfg := embeddedConfig(t, db, true)
			app := openEmbedded(t, cfg)
			if err := app.UnlockDatabase(key); err != nil {
				t.Fatal(err)
			}
			credentials := `{"username":"synthetic-fhir-models","password":"synthetic-password","join_mailing_list":false}`
			signup := embeddedRequest(t, app, "POST", "/api/auth/signup", credentials)
			if signup.Code != 200 {
				t.Fatalf("signup status %d", signup.Code)
			}
			var auth struct {
				Data string `json:"data"`
			}
			if err := json.Unmarshal(signup.Body.Bytes(), &auth); err != nil {
				t.Fatal(err)
			}
			original := []byte(fmt.Sprintf(`{"resourceType":"Patient","id":"synthetic-model-patient","gender":%q,"active":false,"multipleBirthInteger":0,"birthDate":"2000-02","meta":{"lastUpdated":"2024-02-29T23:59:59.123456+03:00"},"name":[{"text":"Synthetic عربي"}],"extension":[{"url":"https://example.invalid/false","valueBoolean":false}],"syntheticUnknown":{"integer":9007199254740993,"values":[null,false,0]}}`, gender))
			var body bytes.Buffer
			form := multipart.NewWriter(&body)
			part, err := form.CreateFormFile("file", "synthetic-model-import.json")
			if err != nil {
				t.Fatal(err)
			}
			if _, err = fmt.Fprintf(part, `{"resourceType":"Bundle","type":"collection","entry":[{"resource":%s}]}`, original); err != nil {
				t.Fatal(err)
			}
			if err = form.Close(); err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("POST", "/api/secure/source/manual", &body)
			req.Header.Set("Content-Type", form.FormDataContentType())
			req.Header.Set("Authorization", "Bearer "+auth.Data)
			response := httptest.NewRecorder()
			app.ServeHTTP(response, req)
			if response.Code != 200 {
				t.Fatalf("synthetic %s import status %d: %s", gender, response.Code, response.Body.String())
			}
			check := func(app *EmbeddedApplication, token string) {
				t.Helper()
				req := httptest.NewRequest("GET", "/api/secure/resource/fhir?sourceResourceType=Patient", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				response := httptest.NewRecorder()
				app.ServeHTTP(response, req)
				if response.Code != 200 {
					t.Fatalf("query status %d", response.Code)
				}
				var result struct {
					Data []struct {
						Raw json.RawMessage `json:"resource_raw"`
					} `json:"data"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				if len(result.Data) != 1 {
					t.Fatalf("patient count %d", len(result.Data))
				}
				if !reflect.DeepEqual(strictFHIRObject(t, original), strictFHIRObject(t, result.Data[0].Raw)) {
					t.Fatal("raw FHIR payload changed, including shorthand/zero/false/unknown precision/date/unicode")
				}
			}
			check(app, auth.Data)
			if err := app.Close(); err != nil {
				t.Fatal(err)
			}
			reopened := openEmbedded(t, embeddedConfig(t, db, true))
			if reopened.UnlockDatabase("wrong-synthetic-key") == nil {
				t.Fatal("wrong key accepted")
			}
			if err := reopened.UnlockDatabase(key); err != nil {
				t.Fatal(err)
			}
			signin := embeddedRequest(t, reopened, "POST", "/api/auth/signin", credentials)
			if signin.Code != 200 {
				t.Fatalf("reopened signin status %d", signin.Code)
			}
			if err := json.Unmarshal(signin.Body.Bytes(), &auth); err != nil {
				t.Fatal(err)
			}
			check(reopened, auth.Data)
			if err := reopened.Close(); err != nil {
				t.Fatal(err)
			}
			payload, err := os.ReadFile(db)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.HasPrefix(payload, []byte("SQLite format 3")) {
				t.Fatal("fixture is not encrypted")
			}
			if blocked.requests != 0 {
				t.Fatal("local synthetic import attempted network")
			}
			leftovers, err := filepath.Glob(filepath.Join(os.Getenv("TMPDIR"), "synthetic-model-import.json*"))
			if err != nil {
				t.Fatal(err)
			}
			if len(leftovers) != 0 {
				t.Fatal("temporary synthetic upload retained")
			}
		})
	}
}
