.PHONY: all clean

all:
	go build -o ./bin/sprig ./cmd/main.go

clean:
	rm -f ./bin/sprig
