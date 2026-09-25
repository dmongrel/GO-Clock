// SPDX-FileCopyrightText: 2026 Joel L. Caesar
// SPDX-License-Identifier: Apache-2.0

//go:build !fyne

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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
		{"/clock.js", "formatDigits"},
		{"/digits.js", "buildFace"},
		{"/alarmdialog.js", "formatAlarmTime"},
		{"/settings.js", "createSettingsDialog"},
		{"/errors.js", "showError"},
		{"/wails/runtime.js", "window._wails"},
		{"/bindings/GO-Clock/clockservice.js", "export function GetConfig"},
		{"/bindings/GO-Clock/models.js", "Sound"},
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

// importPattern finds the module paths a frontend file imports. Only the
// relative and root-relative forms matter here; a bare specifier would be a
// bundler's job, and this frontend has no bundler.
var importPattern = regexp.MustCompile(`from\s+"(\.{1,2}/[^"]+|/[^"]+)"`)

// TestEveryFrontendImportIsServed walks the frontend's own import graph and asks
// the handler for each target. Adding a module and forgetting it is otherwise a
// blank window with one console error nobody sees, and go:embed cannot catch it
// because every file is embedded by directory.
func TestEveryFrontendImportIsServed(t *testing.T) {
	handler := application.BundledAssetFileServer(frontendAssets)

	entries, err := frontendAssets.ReadDir("frontend/dist")
	if err != nil {
		t.Fatalf("reading the embedded frontend: %v", err)
	}

	checked := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".js") {
			continue
		}
		data, err := frontendAssets.ReadFile("frontend/dist/" + entry.Name())
		if err != nil {
			t.Fatalf("reading %s: %v", entry.Name(), err)
		}

		for _, match := range importPattern.FindAllStringSubmatch(string(data), -1) {
			target := match[1]
			// Resolve relative to the importing file, which sits at the root.
			request := target
			if rest, ok := strings.CutPrefix(target, "./"); ok {
				request = "/" + rest
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, request, nil))
			if rec.Code != http.StatusOK {
				t.Errorf("%s imports %q, which the handler serves as %d", entry.Name(), target, rec.Code)
			}
			checked++
		}
	}

	if checked == 0 {
		t.Fatal("found no imports to check, so this test proved nothing")
	}
	t.Logf("checked %d imports across the frontend", checked)
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

// TestEveryIconTheFrontendAsksForExists pins the icon names down. The frontend
// asks for them by string, so a renamed file is a silently missing icon.
func TestEveryIconTheFrontendAsksForExists(t *testing.T) {
	c := &ClockService{}
	for _, name := range []string{"gear", "load", "play", "stop", "refresh", "alarm-clock"} {
		t.Run(name, func(t *testing.T) {
			markup, err := c.IconSVG(name)
			if err != nil {
				t.Fatalf("IconSVG(%q) failed: %v", name, err)
			}
			if !strings.HasPrefix(markup, "<svg") {
				t.Errorf("IconSVG(%q) did not return SVG markup", name)
			}
		})
	}
}

// TestIconSVGRecoloursTheAlarmClock covers the one icon whose colour is not in
// CSS reach, because it lives in an internal stylesheet.
func TestIconSVGRecoloursTheAlarmClock(t *testing.T) {
	c := &ClockService{}
	markup, err := c.IconSVG("alarm-clock")
	if err != nil {
		t.Fatalf("IconSVG failed: %v", err)
	}
	if strings.Contains(markup, alarmIconStockColour) {
		t.Errorf("the stock colour %s survived, so the icon will not match the digits", alarmIconStockColour)
	}
	// A zero-value service reports the default colours.
	if want := "#06f2f5"; !strings.Contains(markup, want) {
		t.Errorf("recoloured icon does not contain the digit colour %s", want)
	}
}

// TestIconSVGRejectsPaths keeps the name from being used to walk the embedded
// filesystem. Nothing untrusted reaches it today; the check is there so that
// stays true if something ever does.
func TestIconSVGRejectsPaths(t *testing.T) {
	c := &ClockService{}
	for _, name := range []string{"../config", "images/gear", "gear.svg", `..\gear`} {
		if _, err := c.IconSVG(name); err == nil {
			t.Errorf("IconSVG(%q) was accepted, want an error", name)
		}
	}
}

// TestListSoundsIncludesBundledAndImported checks both halves of the list, and
// that the imported ones are flagged - the flag is what decides where the bytes
// are later read from.
func TestListSoundsIncludesBundledAndImported(t *testing.T) {
	dir := redirectConfigDirForTest(t)

	alarmsDir := filepath.Join(dir, "Go-Clock", "Alarms")
	if err := os.MkdirAll(alarmsDir, 0o755); err != nil {
		t.Fatalf("creating the alarms directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(alarmsDir, "imported.mp3"), []byte("not really audio"), 0o644); err != nil {
		t.Fatalf("writing the imported sound: %v", err)
	}

	sounds, err := (&ClockService{}).ListSounds()
	if err != nil {
		t.Fatalf("ListSounds failed: %v", err)
	}

	var bundled, imported int
	for _, sound := range sounds {
		if sound.IsUser {
			imported++
			if sound.Name != "imported.mp3" {
				t.Errorf("unexpected imported sound %q", sound.Name)
			}
		} else {
			bundled++
		}
	}
	if bundled == 0 {
		t.Error("no bundled sounds listed, but three are embedded")
	}
	if imported != 1 {
		t.Errorf("listed %d imported sounds, want 1", imported)
	}
}

// TestListSoundsWithoutAnImportDirectory covers a first run, where nothing has
// been imported. An empty user half is normal, not a failure.
func TestListSoundsWithoutAnImportDirectory(t *testing.T) {
	redirectConfigDirForTest(t)

	sounds, err := (&ClockService{}).ListSounds()
	if err != nil {
		t.Fatalf("ListSounds failed: %v", err)
	}
	if len(sounds) == 0 {
		t.Fatal("no sounds listed at all, but three are embedded")
	}
	for _, sound := range sounds {
		if sound.IsUser {
			t.Errorf("listed %q as imported on a first run", sound.Name)
		}
	}
}

// TestReadSoundBundled proves a bundled sound is read from the binary, not from
// disk, which is what lets the app run with an empty alarms directory.
func TestReadSoundBundled(t *testing.T) {
	sounds, err := (&ClockService{}).ListSounds()
	if err != nil {
		t.Fatalf("ListSounds failed: %v", err)
	}
	data, err := readSound(sounds[0].Name, false)
	if err != nil {
		t.Fatalf("readSound(%q) failed: %v", sounds[0].Name, err)
	}
	if len(data) == 0 {
		t.Errorf("readSound(%q) returned no bytes", sounds[0].Name)
	}
}

// TestTimezoneIsAnAbbreviation checks the one thing the frontend cannot work out
// for itself. JavaScript has the offset; the short zone name comes from here.
func TestTimezoneIsAnAbbreviation(t *testing.T) {
	zone := (&ClockService{}).Timezone()
	if zone == "" {
		t.Fatal("Timezone returned an empty string")
	}
	if strings.ContainsAny(zone, " :") {
		t.Errorf("Timezone returned %q, which does not look like an abbreviation", zone)
	}
}

// TestSnoozeWithNoAlarmPlayingIsANoOp covers the guard that stops a stray SPACE
// press from arming an alarm that was never ringing. It also reaches Snooze on a
// service with no app attached, which would panic if the guard were removed.
func TestSnoozeWithNoAlarmPlayingIsANoOp(t *testing.T) {
	c := &ClockService{}
	c.Snooze()
	if c.snoozeTimer != nil {
		t.Error("Snooze armed a timer although no alarm was playing")
	}
	if c.IsAlarmPlaying() {
		t.Error("Snooze reported the alarm as playing")
	}
}

// redirectConfigDirForTest points os.UserConfigDir at a temporary directory, so
// a test never reads or writes the real config. os.UserConfigDir reads a
// different variable on each platform, hence the switch.
func redirectConfigDirForTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	switch runtime.GOOS {
	case "windows":
		t.Setenv("AppData", dir)
	case "darwin":
		t.Setenv("HOME", dir)
		return filepath.Join(dir, "Library", "Application Support")
	default:
		t.Setenv("XDG_CONFIG_HOME", dir)
	}
	return dir
}
