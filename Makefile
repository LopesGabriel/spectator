.PHONY: build clean

build:
	@echo "Building Go application..."
	GOOS=windows GOARCH=amd64	\
	CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc \
	go build -o out/consumer.exe cmd/consumer/main.go

clean:
	@echo "Cleaning build artifacts..."
	rm -rf out/consumer.exe