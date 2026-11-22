
BIN := bin/chat

.PHONY: all build clean

all: build

build: $(BIN)

$(BIN): main.go
	mkdir -p $(@D)
	go build -o $@ main.go

clean:
	rm -rf bin
