.PHONY: build
build:
	go build -o bin/master master/main/master.go
	go build -o bin/worker worker/main/worker.go

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: clean
clean:
	rm -rf bin
