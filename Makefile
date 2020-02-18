asteroids:
	go generate
	go build -ldflags="-s -w"
	upx asteroids

clean:
	rm -f asteroids bindata.go
