package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shop/backend/internal/domain"
	"github.com/shop/backend/internal/transport/http/dto"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleError_ValidationError(t *testing.T) {
	c, recorder := newContext()
	handleError(c, &domain.ValidationError{Field: "price", Message: "price must be greater than zero"})
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	body := decodeError(t, recorder)
	if body.Error.Code != "validation" {
		t.Errorf("code = %q, want %q", body.Error.Code, "validation")
	}
	if body.Error.Message != "validation failed" {
		t.Errorf("message = %q, want %q", body.Error.Message, "validation failed")
	}
	if len(body.Error.Details) != 1 || body.Error.Details[0].Field != "price" || body.Error.Details[0].Message != "price must be greater than zero" {
		t.Errorf("details = %+v, want one detail with field=price", body.Error.Details)
	}
	if body.Error.Field != nil {
		t.Errorf("field must be omitted when details are present, got %v", *body.Error.Field)
	}
}

func TestHandleError_Sentinels(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantMsg    string
	}{
		{
			name:       "not found",
			err:        domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found",
			wantMsg:    "entity not found",
		},
		{
			name:       "already exists",
			err:        domain.ErrAlreadyExists,
			wantStatus: http.StatusConflict,
			wantCode:   "already_exists",
			wantMsg:    "entity already exists",
		},
		{
			name:       "in use",
			err:        domain.ErrInUse,
			wantStatus: http.StatusConflict,
			wantCode:   "in_use",
			wantMsg:    "entity is in use",
		},
		{
			name:       "wrapped not found is still recognized",
			err:        errors.Join(errors.New("wrapped"), domain.ErrNotFound),
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found",
			wantMsg:    "entity not found",
		},
		{
			name:       "unknown error becomes internal",
			err:        errors.New("some database timeout"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "internal",
			wantMsg:    "internal server error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, recorder := newContext()
			handleError(c, tt.err)
			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			body := decodeError(t, recorder)
			if body.Error.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", body.Error.Code, tt.wantCode)
			}
			if body.Error.Message != tt.wantMsg {
				t.Errorf("message = %q, want %q", body.Error.Message, tt.wantMsg)
			}
			if body.Error.Field != nil {
				t.Errorf("field must be omitted for sentinels, got %v", *body.Error.Field)
			}
			if len(body.Error.Details) != 0 {
				t.Errorf("details must be empty for sentinels, got %v", body.Error.Details)
			}
		})
	}
}

func TestHandleBindingError_ValidatorErrors(t *testing.T) {
	type sample struct {
		Name  string `json:"name" binding:"required"`
		Price int    `json:"price" binding:"required,gt=0"`
	}
	v := validator.New()
	err := v.Struct(sample{})
	c, recorder := newContext()
	handleBindingError(c, err)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	body := decodeError(t, recorder)
	if body.Error.Code != "validation" {
		t.Errorf("code = %q, want %q", body.Error.Code, "validation")
	}
	if body.Error.Message != "validation failed" {
		t.Errorf("message = %q, want %q", body.Error.Message, "validation failed")
	}
	if len(body.Error.Details) != 2 {
		t.Fatalf("len(details) = %d, want 2", len(body.Error.Details))
	}
	fields := map[string]string{}
	for _, d := range body.Error.Details {
		fields[d.Field] = d.Message
	}
	if fields["Name"] != "required" {
		t.Errorf("Name detail = %q, want required", fields["Name"])
	}
	if fields["Price"] != "required" {
		t.Errorf("Price detail = %q, want required (first failing tag)", fields["Price"])
	}
}

func TestHandleBindingError_OtherError(t *testing.T) {
	c, recorder := newContext()
	handleBindingError(c, errors.New("unexpected json"))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	body := decodeError(t, recorder)
	if body.Error.Code != "validation" {
		t.Errorf("code = %q, want %q", body.Error.Code, "validation")
	}
	if body.Error.Message != "invalid request payload" {
		t.Errorf("message = %q, want %q", body.Error.Message, "invalid request payload")
	}
}

func TestParseUintParam_Valid(t *testing.T) {
	c, _ := newContext()
	c.Params = gin.Params{{Key: "id", Value: "42"}}
	id, ok := parseUintParam(c, "id")
	if !ok || id != 42 {
		t.Fatalf("parseUintParam = (%d, %v), want (42, true)", id, ok)
	}
}

func TestParseUintParam_Missing(t *testing.T) {
	c, recorder := newContext()
	id, ok := parseUintParam(c, "id")
	if ok {
		t.Fatalf("parseUintParam = (_, true), want false on missing param")
	}
	if id != 0 {
		t.Errorf("returned id = %d, want 0", id)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	body := decodeError(t, recorder)
	if body.Error.Code != "validation" || body.Error.Field == nil || *body.Error.Field != "id" {
		t.Errorf("body = %+v, want validation error with field=id", body)
	}
}

func TestParseUintParam_Invalid(t *testing.T) {
	c, recorder := newContext()
	c.Params = gin.Params{{Key: "id", Value: "not-a-number"}}
	id, ok := parseUintParam(c, "id")
	if ok {
		t.Fatalf("parseUintParam = (_, true), want false on invalid")
	}
	if id != 0 {
		t.Errorf("returned id = %d, want 0", id)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	body := decodeError(t, recorder)
	if body.Error.Message != "invalid id" {
		t.Errorf("message = %q, want %q", body.Error.Message, "invalid id")
	}
	if body.Error.Field == nil || *body.Error.Field != "id" {
		t.Errorf("field = %v, want id", body.Error.Field)
	}
}

func newContext() (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, recorder
}

func decodeError(t *testing.T, recorder *httptest.ResponseRecorder) dto.ErrorResponse {
	t.Helper()
	var body dto.ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v, body: %s", err, recorder.Body.String())
	}
	return body
}