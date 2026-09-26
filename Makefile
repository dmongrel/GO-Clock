.PHONY: build dev bindings ico run test dist snapshot

# -tags production swaps the bundled Wails runtime from the 516 KB debug module
# to the 57 KB one and drops the devtools, so it is what a release wants.
build: ico bindings
	GOOS=windows GOARCH=amd64 go build -tags production -ldflags="-H=windowsgui" -o Go-Clock.exe

# A development build: the debug runtime, the devtools, and a console window, so
# a frontend error is visible instead of silent.
dev: bindings
	go build -o Go-Clock-dev.exe

# The frontend calls into Go through generated bindings. -b bundles the runtime
# rather than expecting an npm @wailsio/runtime install, which is what lets this
# frontend stay plain files with no node toolchain.
bindings:
	wails3 generate bindings -b -d frontend/dist/bindings

ico:
	rsrc -manifest app.manifest -ico Go-Clock.ico -o ico.syso

run: build
	./Go-Clock.exe

test:
	go test ./... -count=1

dist: build
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean
