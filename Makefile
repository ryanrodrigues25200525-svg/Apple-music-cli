.PHONY: build run clean test install

GO_BIN ?= /opt/homebrew/bin/go
BINARY_NAME = mu
VERSION ?= dev
LDFLAGS = -X github.com/ryanrodrigues25200525-svg/Apple-music-cli/cmd.Version=$(VERSION)

build:
	$(GO_BIN) build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) main.go

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

test:
	$(GO_BIN) test ./...

install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY_NAME) $(HOME)/.local/bin/$(BINARY_NAME)
