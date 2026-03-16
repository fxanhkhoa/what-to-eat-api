package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// --- CustomValidator tests ---

type validStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=0,lte=130"`
}

type invalidStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func newCustomValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

func TestCustomValidator_ValidStruct(t *testing.T) {
	cv := newCustomValidator()
	s := validStruct{Name: "Alice", Email: "alice@example.com", Age: 25}
	if err := cv.Validate(s); err != nil {
		t.Errorf("expected no validation error, got: %v", err)
	}
}

func TestCustomValidator_MissingRequired(t *testing.T) {
	cv := newCustomValidator()
	s := invalidStruct{Name: "", Email: ""}
	if err := cv.Validate(s); err == nil {
		t.Error("expected validation error for missing required fields, got nil")
	}
}

func TestCustomValidator_InvalidEmail(t *testing.T) {
	cv := newCustomValidator()
	s := invalidStruct{Name: "Bob", Email: "not-an-email"}
	if err := cv.Validate(s); err == nil {
		t.Error("expected validation error for invalid email, got nil")
	}
}

func TestCustomValidator_ValidEmail(t *testing.T) {
	cv := newCustomValidator()
	s := invalidStruct{Name: "Bob", Email: "bob@example.com"}
	if err := cv.Validate(s); err != nil {
		t.Errorf("expected no validation error, got: %v", err)
	}
}

func TestCustomValidator_AgeOutOfRange(t *testing.T) {
	cv := newCustomValidator()
	s := validStruct{Name: "Charlie", Email: "charlie@example.com", Age: 200}
	if err := cv.Validate(s); err == nil {
		t.Error("expected validation error for age out of range, got nil")
	}
}

func TestCustomValidator_NonStruct(t *testing.T) {
	cv := newCustomValidator()
	// Passing a primitive — validator should return an error
	if err := cv.Validate("plain string"); err == nil {
		t.Error("expected validation error when passing non-struct, got nil")
	}
}

// --- Root handler test ---

func TestRootHandler(t *testing.T) {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	body := strings.TrimSpace(rec.Body.String())
	if body != "OK" {
		t.Errorf("expected body 'OK', got %q", body)
	}
}
