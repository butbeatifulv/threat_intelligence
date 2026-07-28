package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteJSON(rr, http.StatusOK, map[string]any{"x": 1})
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["x"].(float64) != 1 {
		t.Fatalf("body %v", body)
	}
}

func TestWriteError_prod(t *testing.T) {
	SetProdMode(true)
	defer SetProdMode(false)

	rr := httptest.NewRecorder()
	WriteError(rr, http.StatusBadRequest, errors.New("secret detail"))
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["error"] != "bad request" {
		t.Fatalf("got %v", body["error"])
	}

	rr2 := httptest.NewRecorder()
	WriteError(rr2, http.StatusNotFound, errors.New("missing node"))
	_ = json.Unmarshal(rr2.Body.Bytes(), &body)
	if body["error"] != "not found" {
		t.Fatalf("got %v", body["error"])
	}
}

func TestSanitizeError_dev(t *testing.T) {
	SetProdMode(false)
	if got := SanitizeError(http.StatusBadRequest, "detail"); got != "detail" {
		t.Fatalf("got %q", got)
	}
}

func TestSanitizeError_prod_internal(t *testing.T) {
	SetProdMode(true)
	defer SetProdMode(false)
	if got := SanitizeError(http.StatusInternalServerError, "db timeout"); got != "internal error" {
		t.Fatalf("got %q", got)
	}
}
