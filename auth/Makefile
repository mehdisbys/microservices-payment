lint:
		golangci-lint run --config=.golangci.yml ./...

test:	lint
		go test -cover -failfast ./...

linux-binary:
		GOOS=linux GOARCH=amd64 go build -a -o main -ldflags "-w -s"