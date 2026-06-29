.PHONY: build ico run dist

build: ico
	GOOS=windows GOARCH=amd64 go build -ldflags="-H=windowsgui" -o Go-Clock.exe

ico:
	rsrc -manifest app.manifest -ico Go-Clock.ico -o ico.syso

run: build
	./Go-Clock.exe

dist: build
	powershell -Command "Compress-Archive -Path Go-Clock.exe -DestinationPath dist/Go-Clock_1.0.0_Windows_x64.zip -Force"
