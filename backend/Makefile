.PHONY: build test fmt clean

build: fmt
	go mod vendor
	go mod tidy
	go build -o starter ./cmd/starter
	go build -o cmder ./cmd/cmder
	go build -o bootstrap ./cmd/raft
	go build -o integrate_test/starter ./cmd/starter

integrate: build
	sh ./integrate_test/start_cluster.sh

test:
	go test -v ./...

fmt:
	go fmt ./...
	go vet ./...

clean:
	rm -f starter cmder bootstrap integrate_test/starter

