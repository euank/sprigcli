.PHONY: all clean functional_tests release

all: clean
	goreleaser release

clean:
	rm -rf dist/

functional_tests: all
	go test -v ./tests/...
