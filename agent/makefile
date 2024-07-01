APP_NAME := idbagent
VERSION := 1.0.0
BINARY_LINUX := $(APP_NAME)-linux
BINARY_MACOS := $(APP_NAME)-macos

.PHONY: all linux macos clean install-linux install-macos

all: linux macos

linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_LINUX) -ldflags "-X main.version=$(VERSION)"

macos:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_MACOS) -ldflags "-X main.version=$(VERSION)"

clean:
	rm -f $(BINARY_LINUX) $(BINARY_MACOS)

install-linux: linux
	sudo mv $(BINARY_LINUX) /usr/local/bin/$(APP_NAME)
	sudo chmod +x /usr/local/bin/$(APP_NAME)

install-macos: macos
	sudo mv $(BINARY_MACOS) /usr/local/bin/$(APP_NAME)
	sudo chmod +x /usr/local/bin/$(APP_NAME)
