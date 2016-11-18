package errors

import (
	"encoding/json"
	"reflect"
	"testing"
)

func toMap(t *testing.T, v interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	byt, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(byt, &m)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestStackTrace(t *testing.T) {
	err := ErrInternalServerError.New(0)
	if err.StackTrace == nil {
		t.Error("ErrInternalServerError.New(0): StackTrace is nil")
	}

	err = ErrInternalServerError.New(400)
	if err.StackTrace != nil {
		t.Error("ErrInternalServerError.New(400): StackTrace is not nil")
	}

	err = ErrInternalServerError.NewWithStack(400, 0)
	if err.StackTrace == nil {
		t.Error("ErrInternalServerError.NewWithStack(400): StackTrace is nil")
	}
}

func TestErrorAsJSON(t *testing.T) {
	err := ErrInternalServerError.New(0)
	js, _ := err.AsJSON()
	m := make(map[string]interface{})
	e := json.Unmarshal([]byte(js), &m)
	if e != nil {
		t.Errorf("AsJSON(): Invalid JSON: %s", e)
	}

	if len(m) != 7 {
		t.Error("More than 7 keys in json from AsJSON")
	}

	for _, k := range []string{
		"id",
		"class",
		"details",
		"internal_error",
		"internal_details",
		"stack_trace",
		"status",
	} {
		if _, ok := m[k]; !ok {
			t.Errorf("Missing key '%s' from AsJSON", k)
		}
	}

	_, ok := m["status"].(float64)
	if !ok {
		t.Errorf("status is not a number, type is: %s",
			reflect.TypeOf(m["status"]),
		)
	}
}

func TestErrorJSONAPI(t *testing.T) {
	err := ErrInternalServerError.New(0)
	jsonapi_err := err.AsJSONAPIError()

	m := toMap(t, jsonapi_err)

	if len(m) != 4 {
		t.Error("More than 4 keys in jsonapi errors container")
	}

	for _, k := range []string{"id", "code", "status", "title"} {
		if _, ok := m[k]; !ok {
			t.Errorf("Missing key '%s' in JSONAPIError", k)
		}
	}

	_, ok := m["status"].(string)
	if !ok {
		t.Errorf("status is not string")
	}
}

func TestErrorsJSONAPI(t *testing.T) {
	errs := make(Errors, 0, 2)
	errs.AddError(ErrInternalServerError.New(0))
	errs.AddError(ErrJSONSchemaValidationFailed.New(0))

	jsonapi_errs := errs.AsJSONAPIResponse()

	m := toMap(t, jsonapi_errs)
	if len(m) != 1 {
		t.Error("More than 1 key in jsonapi errors container")
	}

	errors := m["errors"].([]interface{})
	if errors == nil {
		t.Error("No 'errors' key in top level map")
	}

	if len(errors) != 2 {
		t.Errorf("Length of errors (%d) != 2", len(errors))
	}

	err := errors[0].(map[string]interface{})
	if err["code"].(string) != ErrInternalServerError.Code {
		t.Errorf("First error is (%+v), not ErrInternalServerError", err)
	}

	err = errors[1].(map[string]interface{})
	if err["code"].(string) != ErrJSONSchemaValidationFailed.Code {
		t.Errorf("First error is (%+v), not ErrJSONSchemaValidationFailed", err)
	}
}

func TestErrorsJSON(t *testing.T) {
	errs := make(Errors, 0, 2)
	errs.AddError(ErrInternalServerError.New(0))
	errs.AddError(ErrJSONSchemaValidationFailed.New(0))

	jsonapi_errs := errs.AsJSONAPIResponse()

	m := toMap(t, jsonapi_errs)
	if len(m) != 1 {
		t.Error("More than 1 key in jsonapi errors container")
	}

	errors := m["errors"].([]interface{})
	if errors == nil {
		t.Error("No 'errors' key in top level map")
	}

	if len(errors) != 2 {
		t.Errorf("Length of errors (%d) != 2", len(errors))
	}

	err := errors[0].(map[string]interface{})
	if err["code"].(string) != ErrInternalServerError.Code {
		t.Errorf("First error is (%+v), not ErrInternalServerError", err)
	}

	err = errors[1].(map[string]interface{})
	if err["code"].(string) != ErrJSONSchemaValidationFailed.Code {
		t.Errorf("First error is (%+v), not ErrJSONSchemaValidationFailed", err)
	}
}
