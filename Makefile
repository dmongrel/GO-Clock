.PHONY: build package run dev bindings icon test

# Thin wrappers over the Wails3 build scaffold in Taskfile.yml and build/.
# Everything the build actually does lives there; this file exists so `make`
# still works from muscle memory.

# A production build: bin/Go-Clock.exe, the small runtime, no console window.
build:
	wails3 task build

# A release: the production build wrapped in an NSIS installer, written to
# bin/Go-Clock-amd64-installer.exe.
package:
	wails3 task package

run:
	wails3 task run

# A development build: the debug runtime, the devtools, and a console window,
# so a frontend error is visible instead of silent.
dev:
	wails3 task build DEV=true

bindings:
	wails3 task common:generate:bindings

# Only needed after images/alarm-clock.svg changes: SVG -> build/appicon.png
# -> build/windows/icon.ico.
icon:
	go run ./scripts
	wails3 task common:generate:icons

test:
	go test ./... -count=1
