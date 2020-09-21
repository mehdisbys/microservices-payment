.PHONY: all test

deps:
	export GO111MODULE=on ; go mod download

all:
	make -C ./payment linux-binary
	make -C ./gateway linux-binary
	make -C ./auth linux-binary

test:
	make -C ./payment test-dependencies
	make -C ./payment test
	make -C ./payment test-clean
	make -C ./gateway test
	make -C ./auth test



clean:
	rm -f ./payment/main
	rm -f ./gateway/main
	rm -f ./auth/main

deploy:
	docker-compose -f docker-compose.yaml up