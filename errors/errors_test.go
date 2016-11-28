package errors

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

type E1 struct{}

func (self *E1) Error() string { return "woot1" }

type E2 struct{}

func (self *E2) String() string { return "woot2" }

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

func TestIncludedErrors(t *testing.T) {
	if len(defaultErrorManager.errorClasses) != 4 {
		t.Errorf(
			".errorClasses is not 4: %+v",
			defaultErrorManager.errorClasses,
		)
	}
	for k, _ := range defaultErrorManager.errorClasses {
		idx := strings.LastIndex(k, ".")
		name := k[idx+1:]
		if name != "ErrInternalServerError" &&
			name != "ErrJSONSchemaValidationFailed" &&
			name != "ErrRouteNotFound" &&
			name != "ErrInternalError" {
			t.Errorf("Unexpected error found in .errorClasses: %s", k)
		}
	}
}

func TestErrorClasses(t *testing.T) {
	classes := defaultErrorManager.ErrorClasses()
	found := make(map[string]bool)
	for _, cls := range classes {
		found[cls.Name] = true
	}
	if len(found) != 4 {
		t.Errorf(
			"ErrorClasses() didn't return 4 unique error classes: %+v",
			classes,
		)
	}
	for k, _ := range found {
		idx := strings.LastIndex(k, ".")
		name := k[idx+1:]
		if name != "ErrInternalServerError" &&
			name != "ErrJSONSchemaValidationFailed" &&
			name != "ErrRouteNotFound" &&
			name != "ErrInternalError" {
			t.Errorf("Unexpected error found in ErrorClasses(): %s", k)
		}
	}
}

func TestInterface(t *testing.T) {
	var err interface{} = &Error{}
	var errs interface{} = &Errors{}

	if _, ok := err.(ErrorType); !ok {
		t.Error("Error does not satisfy ErrorType")
	}

	if _, ok := errs.(ErrorType); !ok {
		t.Error("Errors does not satisfy ErrorType")
	}
}

func TestStackTrace(t *testing.T) {
	err := ErrInternalServerError.New("")
	if err.StackTrace == nil {
		t.Error("ErrInternalServerError.New: StackTrace is nil")
	}

	err = ErrJSONSchemaValidationFailed.New("")
	if err.StackTrace != nil {
		t.Error("ErrJSONSchemaValidationFailed.New(): StackTrace is not nil")
	}

	err = ErrJSONSchemaValidationFailed.NewWithStack("", 0)
	if err.StackTrace == nil {
		t.Error("ErrJSONSchemaValidationFailed.NewWithStack(): StackTrace is nil")
	}
}

func TestNewWithDetails(t *testing.T) {
	err := ErrInternalServerError.New("test1")
	if err.Details != "test1" {
		t.Error("Details not set correctly")
	}

	err = ErrInternalServerError.NewWithStack("test2", 0)
	if err.Details != "test2" {
		t.Error("Details not set correctly")
	}
}

func TestSetInternal(t *testing.T) {
	err := ErrInternalServerError.New("").SetInternal("test")
	if err.InternalDetails.(string) != "test" {
		t.Error("InternalDetails not set correctly")
	}
	if err.InternalError != "test" {
		t.Error("InternalError not set correctly")
	}

	err = ErrInternalServerError.NewWithStack("", 0).SetInternal("test")
	if err.InternalDetails.(string) != "test" {
		t.Error("InternalDetails not set correctly")
	}
	if err.InternalError != "test" {
		t.Error("InternalError not set correctly")
	}

	e1 := &E1{}
	e2 := &E2{}

	err = ErrInternalServerError.New("").SetInternal(e1)
	if err.InternalDetails.(*E1) != e1 {
		t.Error("InternalDetails not set correctly from e1")
	}
	if err.InternalError != "woot1" {
		t.Error("InternalError not set correctly from e1.Error()")
	}

	err = ErrInternalServerError.New("").SetInternal(e2)
	if err.InternalDetails.(*E2) != e2 {
		t.Error("InternalDetails not set correctly from e2")
	}
	if err.InternalError != "woot2" {
		t.Error("InternalError not set correctly from e2.String()")
	}
}

func TestErrorAsJSON(t *testing.T) {
	err := ErrInternalServerError.New(
		"details",
	).SetInternal(
		"internal",
	).SetMetadata(
		make(map[string]interface{}),
	).SetInternalMetadata(
		make(map[string]interface{}),
	)
	js, _ := err.AsJSON()
	m := make(map[string]interface{})
	e := json.Unmarshal([]byte(js), &m)
	if e != nil {
		t.Errorf("AsJSON(): Invalid JSON: %s", e)
	}

	if len(m) != 7 {
		t.Errorf("Didn't find 7 keys in json from AsJSON: %+v", m)
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
	err := ErrInternalServerError.New("details")
	jsonapi_err := err.AsJSONAPIError()

	m := toMap(t, jsonapi_err)

	if len(m) != 5 {
		t.Error("Didn't find 5 keys in jsonapi errors container")
	}

	for _, k := range []string{"id", "code", "status", "detail", "title"} {
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
	errs.AddError(ErrInternalServerError.New(""))
	errs.AddError(ErrJSONSchemaValidationFailed.New(""))

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
	errs.AddError(ErrJSONSchemaValidationFailed.New(""))
	errs.AddError(ErrInternalServerError.New(""))

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
	if err["code"].(string) != ErrJSONSchemaValidationFailed.Code {
		t.Errorf("First error is (%+v), not ErrInternalServerError", err)
	}

	err = errors[1].(map[string]interface{})
	if err["code"].(string) != ErrInternalServerError.Code {
		t.Errorf("First error is (%+v), not ErrJSONSchemaValidationFailed", err)
	}

	if errs.GetStatus() != 400 {
		j, _ := errs.AsJSON()
		t.Errorf("Status is not the status of the first error: %s", j)
	}

	errs[0], errs[1] = errs[1], errs[0]

	if errs.GetStatus() != 500 {
		j, _ := errs.AsJSON()
		t.Errorf("Status is not the status of the first error: %s", j)
	}
}
