# parsd - Pars Network Node

.PHONY: build run test clean install

# Build parsd
build:
	go build -o bin/parsd ./cmd/parsd

# Build with GPU support
build-gpu:
	CGO_ENABLED=1 go build -tags gpu -o bin/parsd ./cmd/parsd

# Run parsd (L1 sovereign mode)
run: build
	./bin/parsd --mode=l1 --datadir=~/.pars --warp=true --gpu=true

# Run as L2 rollup
run-l2: build
	./bin/parsd --mode=l2 --datadir=~/.pars --warp=true --gpu=true

# Run local devnet
devnet: build
	./bin/parsd --mode=l1 --datadir=/tmp/pars-devnet --rpc=127.0.0.1:9650 --p2p=0.0.0.0:9651

# Test
test:
	go test -v ./...

# Test with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install parsd globally
install: build
	cp bin/parsd /usr/local/bin/

# Generate protobuf
proto:
	protoc --go_out=. --go-grpc_out=. proto/*.proto

# Docker build
docker:
	docker build -t parsdao/parsd:latest .

# Docker run
docker-run:
	docker run -d -p 9650:9650 -p 9651:9651 parsdao/parsd:latest
