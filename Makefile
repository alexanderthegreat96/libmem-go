.PHONY: setup clean test

setup:
	go run ./cmd/setup

clean:
	rm -rf libmem/deps

test:
	go test -v ./libmem
