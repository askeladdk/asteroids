WINEXE=release/Asteroids.exe
OSXAPP=release/Asteroids.app/Contents/MacOS
OSXEXE=${OSXAPP}/Asteroids

all: osx windows

osx:
	mkdir -p release
	mkdir -p ${OSXAPP}
	cp Info.plist release/Asteroids.app/Contents
	go build -ldflags="-s -w" -o ${OSXEXE}
	upx ${OSXEXE}

windows:
	mkdir -p release
	CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H windowsgui" -o ${WINEXE}
	upx ${WINEXE}

clean:
	rm -f bindata.go
	rm -rf release
