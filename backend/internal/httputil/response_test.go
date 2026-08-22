package httputil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONResponse(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), RequestIDKey, "test-req-id")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	JSON(rec, req, http.StatusOK, map[string]string{"message": "hello"})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}

	if resp.Meta == nil || resp.Meta.RequestID != "test-req-id" {
		t.Errorf("expected meta request_id 'test-req-id', got %+v", resp.Meta)
	}
}

func TestErrorJSONResponse(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	ErrorJSON(rec, req, http.StatusNotFound, "NOT_FOUND", "user not found")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var resp ResponseEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if resp.Success {
		t.Errorf("expected success false")
	}

	if resp.Error == nil || resp.Error.Code != "NOT_FOUND" || resp.Error.Message != "user not found" {
		t.Errorf("expected NOT_FOUND error, got %+v", resp.Error)
	}
}
