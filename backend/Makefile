.PHONY: build test fmt clean

build: fmt test
	go mod vendor
	go mod tidy
	go build -o starter ./cmd/starter
	go build -o cmder ./cmd/cmder
	go build -o bootstrap ./cmd/raft
	go build -o integrate_test/starter ./cmd/starter

integrate: build
	sh ./integrate_test/start_cluster.sh

stop:
	sh ./integrate_test/stop.sh

test:
	GIN_MODE=release go test -v ./...
	GIN_MODE=release go test -cover ./...

fmt:
	go fmt ./...
	go vet ./...

empty:
	@echo "This is an empty target, used for testing purposes only."
	echo > integrate_test/config1.json
	echo > integrate_test/config2.json
	echo > integrate_test/config.json

clean:
	rm -f starter cmder bootstrap integrate_test/starter raft.log  mockservice.log logs/*.log

