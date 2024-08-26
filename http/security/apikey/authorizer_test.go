// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.

package apikey

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiKeyAuthenticationSuccessful(t *testing.T) {
	auth := NewAuthorizer(&Config{Name: "Authorization"}, "testkey")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", "Bearer testkey")

	_, b := auth.Authorize(r, w)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Errorf("Expected no error in authentication, but failed with code %d and body %v", resp.StatusCode, string(body))
	}
	if got, ex := w.Code, http.StatusOK; got != ex {
		t.Errorf("Expected status code %d, got %d", ex, got)
	}
	if got, ex := string(body), ""; got != ex {
		t.Errorf("Expected status code %q, got %q", ex, got)
	}
}

func TestApiKeyAuthenticationError(t *testing.T) {
	auth := NewAuthorizer(&Config{Name: "Authorization"}, "testkey")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Add("Authorization", "Bearer wrongKey")

	_, b := auth.Authorize(r, w)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if b {
		t.Errorf("Expected error in Authentication, but was succesfull with code %d and body %v", resp.StatusCode, string(body))
	}
	if got, ex := w.Code, http.StatusUnauthorized; got != ex {
		t.Errorf("Expected status code %d, got %d", ex, got)
	}
	if got, ex := string(body), "ApiKey not valid\n"; got != ex {
		t.Errorf("Expected error massage %q, got %q", ex, got)
	}
}

func TestApiKeyAuthenticationNoKey(t *testing.T) {
	auth := NewAuthorizer(&Config{Name: "Authorization"}, "testkey")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	_, b := auth.Authorize(r, w)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if b {
		t.Errorf("Expected error in Authentication, but was succesfull with code %d and body %v", resp.StatusCode, string(body))
	}
	if got, ex := w.Code, http.StatusUnauthorized; got != ex {
		t.Errorf("Expected status code %d, got %d", ex, got)
	}
	if got, ex := string(body), "Unauthorized\n"; got != ex {
		t.Errorf("Expected status code %q, got %q", ex, got)
	}
}
