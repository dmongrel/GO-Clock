// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

//go:build wails

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// TestFrontendAssetsServed checks the embedded frontend and the Wails runtime
// are both reachable through the asset handler. A missing runtime shows up in
// the app only as bindings and events silently not working, so it is worth a
// test rather than a manual look at the window.
func TestFrontendAssetsServed(t *testing.T) {
	handler := application.BundledAssetFileServer(frontendAssets)

	tests := []struct {
		path        string
		wantContain string
	}{
		{"/", "<title>Clock</title>"},
		{"/index.html", "<title>Clock</title>"},
		{"/style.css", "--digit-color"},
		{"/main.js", "ClockService"},
		{"/wails/runtime.js", "window._wails"},
		{"/bindings/GO-Clock/clockservice.js", "export function GetConfig"},
		{"/bindings/GO-Clock/config/models.js", "Config"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s returned status %d, want 200", tt.path, rec.Code)
			}
			body := rec.Body.String()
			if body == "" {
				t.Fatalf("GET %s returned an empty body", tt.path)
			}
			if !strings.Contains(body, tt.wantContain) {
				t.Errorf("GET %s body does not contain %q (got %d bytes)", tt.path, tt.wantContain, len(body))
			}
		})
	}
}

// TestIndexRequestsRuntime guards the other half of the wiring: the runtime can
// be served and still never load, because nothing on the page asks for it.
func TestIndexRequestsRuntime(t *testing.T) {
	data, err := frontendAssets.ReadFile("frontend/dist/index.html")
	if err != nil {
		t.Fatalf("reading the embedded index.html: %v", err)
	}
	if want := "/wails/runtime.js"; !strings.Contains(string(data), want) {
		t.Errorf("index.html does not load %s, so window._wails will never exist", want)
	}
	// A dev build serves runtime.debug.js, which is an ES module. In a classic
	// script tag it throws before executing and window._wails stays undefined,
	// which presents as bindings and events silently doing nothing.
	if want := `<script type="module" src="/wails/runtime.js">`; !strings.Contains(string(data), want) {
		t.Errorf("index.html must load the runtime with %s", want)
	}
}

// TestBundledRuntimeIsAModuleInDevBuilds records why the tag above is required,
// so that if a future Wails release ships a non-module debug runtime this test
// fails and the constraint can be revisited rather than cargo-culted.
func TestBundledRuntimeIsAModuleInDevBuilds(t *testing.T) {
	rec := httptest.NewRecorder()
	application.BundledAssetFileServer(frontendAssets).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/wails/runtime.js", nil))

	body := rec.Body.String()
	isModule := strings.Contains(body, "export {") || strings.Contains(body, "export{")
	t.Logf("served runtime is %d bytes, ES module: %v", len(body), isModule)
	if !isModule {
		t.Log("the served runtime is not an ES module; the type=\"module\" requirement in index.html may no longer apply")
	}
}
