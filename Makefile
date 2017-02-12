.PHONY: all clean functional_tests release

all:
	go build -ldflags "-X github.com/euank/sprigcli/cmd/sprig.Version=dirty-$(shell git rev-parse --short HEAD)" -o ./bin/sprig ./cmd/main.go

clean:
	rm -f ./bin/sprig

functional_tests: all
	go test -v ./tests/...

release:
	./scripts/release $(VERSION)
