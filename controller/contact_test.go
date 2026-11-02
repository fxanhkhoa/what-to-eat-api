package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"what-to-eat/be/controller"
	"what-to-eat/be/helper"
	"what-to-eat/be/model"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// mockContactSvc — contactServiceProvider stub
// ---------------------------------------------------------------------------

type mockContactSvc struct {
	findResult    []*model.Contact
	findCount     int64
	findErr       error
	findOneResult *model.Contact
	findOneErr    error
	createResult  *model.Contact
	createErr     error
	updateResult  *model.Contact
	updateErr     error
	removeResult  *model.Contact
	removeErr     error
}

func (m *mockContactSvc) Find(query model.QueryContactDto) ([]*model.Contact, int64, error) {
	return m.findResult, m.findCount, m.findErr
}

func (m *mockContactSvc) FindOne(id string) (*model.Contact, error) {
	return m.findOneResult, m.findOneErr
}

func (m *mockContactSvc) Create(dto model.CreateContactDto) (*model.Contact, error) {
	return m.createResult, m.createErr
}

func (m *mockContactSvc) Update(dto model.UpdateContactDto, profile *model.JwtCustomClaims) (*model.Contact, error) {
	return m.updateResult, m.updateErr
}

func (m *mockContactSvc) Remove(id string, profile *model.JwtCustomClaims) (*model.Contact, error) {
	return m.removeResult, m.removeErr
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newEcho() *echo.Echo {
	return echo.New()
}

func newContactCtrl(svc *mockContactSvc) *controller.ContactController {
	return controller.NewContactControllerWithService(svc)
}

func newTestClaim() *model.JwtCustomClaims {
	return &model.JwtCustomClaims{}
}

// ---------------------------------------------------------------------------
// Find
// ---------------------------------------------------------------------------

func TestContactFind_Success(t *testing.T) {
	contacts := []*model.Contact{{Email: "a@b.com"}, {Email: "c@d.com"}}
	svc := &mockContactSvc{findResult: contacts, findCount: 2}
	e := newEcho()
	req := httptest.NewRequest(http.MethodGet, "/contacts?page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newContactCtrl(svc).Find(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var body helper.PaginationObject
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Count != 2 {
		t.Errorf("expected count 2, got %d", body.Count)
	}
}

func TestContactFind_DefaultPagination(t *testing.T) {
	svc := &mockContactSvc{findResult: []*model.Contact{}, findCount: 0}
	e := newEcho()
	req := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newContactCtrl(svc).Find(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestContactFind_WithKeyword(t *testing.T) {
	svc := &mockContactSvc{
		findResult: []*model.Contact{{Email: "test@x.com"}},
		findCount:  1,
	}
	e := newEcho()
	req := httptest.NewRequest(http.MethodGet, "/contacts?page=1&limit=10&keyword=test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newContactCtrl(svc).Find(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestContactFind_ServiceError(t *testing.T) {
	svc := &mockContactSvc{findErr: errors.New("db error")}
	e := newEcho()
	req := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newContactCtrl(svc).Find(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "db error") {
		t.Errorf("expected body to contain 'db error', got %s", rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// FindOne
// ---------------------------------------------------------------------------

func TestContactFindOne_Success(t *testing.T) {
	contact := &model.Contact{Email: "a@b.com", Name: "Alice"}
	svc := &mockContactSvc{findOneResult: contact}
	e := newEcho()
	req := httptest.NewRequest(http.MethodGet, "/contacts/some-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("some-id")

	if err := newContactCtrl(svc).FindOne(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result model.Contact
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if result.Email != "a@b.com" {
		t.Errorf("expected email 'a@b.com', got %s", result.Email)
	}
}

func TestContactFindOne_ServiceError(t *testing.T) {
	svc := &mockContactSvc{findOneErr: errors.New("not found")}
	e := newEcho()
	req := httptest.NewRequest(http.MethodGet, "/contacts/missing", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("missing")

	if err := newContactCtrl(svc).FindOne(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestContactCreate_Success(t *testing.T) {
	contact := &model.Contact{Email: "new@x.com", Name: "Bob"}
	svc := &mockContactSvc{createResult: contact}
	e := newEcho()
	body := `{"email":"new@x.com","name":"Bob","message":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/contacts", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newContactCtrl(svc).Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result model.Contact
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if result.Email != "new@x.com" {
		t.Errorf("expected email 'new@x.com', got %s", result.Email)
	}
}

func TestContactCreate_BindError(t *testing.T) {
	svc := &mockContactSvc{}
	e := newEcho()
	req := httptest.NewRequest(http.MethodPost, "/contacts", strings.NewReader("{invalid json}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newContactCtrl(svc).Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestContactCreate_ServiceError(t *testing.T) {
	svc := &mockContactSvc{createErr: errors.New("insert failed")}
	e := newEcho()
	body := `{"email":"x@y.com"}`
	req := httptest.NewRequest(http.MethodPost, "/contacts", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := newContactCtrl(svc).Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestContactUpdate_Success(t *testing.T) {
	contact := &model.Contact{Email: "up@x.com", Name: "Updated"}
	svc := &mockContactSvc{updateResult: contact}
	e := newEcho()
	body := `{"email":"up@x.com","name":"Updated","message":"msg"}`
	req := httptest.NewRequest(http.MethodPut, "/contacts/abc123", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("abc123")
	c.Set("CLAIM", newTestClaim())

	if err := newContactCtrl(svc).Update(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestContactUpdate_BindError(t *testing.T) {
	svc := &mockContactSvc{}
	e := newEcho()
	req := httptest.NewRequest(http.MethodPut, "/contacts/abc", strings.NewReader("{invalid}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("abc")
	c.Set("CLAIM", newTestClaim())

	if err := newContactCtrl(svc).Update(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestContactUpdate_ServiceError(t *testing.T) {
	svc := &mockContactSvc{updateErr: errors.New("update failed")}
	e := newEcho()
	body := `{"email":"a@b.com"}`
	req := httptest.NewRequest(http.MethodPut, "/contacts/abc", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("abc")
	c.Set("CLAIM", newTestClaim())

	if err := newContactCtrl(svc).Update(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Remove
// ---------------------------------------------------------------------------

func TestContactRemove_Success(t *testing.T) {
	contact := &model.Contact{Email: "del@x.com", Deleted: true}
	svc := &mockContactSvc{removeResult: contact}
	e := newEcho()
	req := httptest.NewRequest(http.MethodDelete, "/contacts/abc123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("abc123")
	c.Set("CLAIM", newTestClaim())

	if err := newContactCtrl(svc).Remove(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var result model.Contact
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if !result.Deleted {
		t.Error("expected contact to be marked deleted")
	}
}

func TestContactRemove_ServiceError(t *testing.T) {
	svc := &mockContactSvc{removeErr: errors.New("remove failed")}
	e := newEcho()
	req := httptest.NewRequest(http.MethodDelete, "/contacts/abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("abc")
	c.Set("CLAIM", newTestClaim())

	if err := newContactCtrl(svc).Remove(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "remove failed") {
		t.Errorf("expected body to contain 'remove failed', got %s", rec.Body.String())
	}
}
