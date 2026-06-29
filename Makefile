.PHONY: build ico run

build: ico
	GOOS=windows GOARCH=amd64 go build -ldflags="-H=windowsgui" -o Go-Clock.exe

ico:
	rsrc -manifest app.manifest -ico Go-Clock.ico -o ico.syso

run: build
	./Go-Clock.exe
