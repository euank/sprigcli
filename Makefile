.PHONY: all clean functional_tests

all:
	go build -o ./bin/sprig ./cmd/main.go

clean:
	rm -f ./bin/sprig

functional_tests: all
	go test -v ./tests/...
