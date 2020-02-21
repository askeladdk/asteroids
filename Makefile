asteroids:
	go generate
	go build -ldflags="-s -w"
	upx asteroids

windows:
	go generate
	CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w"
	upx asteroids.exe

clean:
	rm -f asteroids asteroids.exe bindata.go
