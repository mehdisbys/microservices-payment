lint:
		golangci-lint run --config=.golangci.yml ./...

test:	lint
		go test -tags integration -cover -failfast ./...

linux-binary:
		GOOS=linux GOARCH=amd64 go build -a -o main -ldflags "-w -s"

# Start docker-compose services, e.g for integration tests
test-dependencies:
	docker-compose -f ./docker-compose.yaml up -d

test-clean:
	docker-compose -f ./docker-compose.yaml down