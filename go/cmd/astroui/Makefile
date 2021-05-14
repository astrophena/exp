GOOS    = "windows"
GOARCH  = "amd64"
LDFLAGS = "-s -w -H windowsgui -buildid="

.PHONY: build clean help

build: ## Build
	@ rsrc -manifest main.manifest
	@ GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -trimpath -ldflags=$(LDFLAGS)
	@ tailscale file cp astroui.exe phryne:

clean: ## Clean
	@ rm *.exe *.syso

help: ## Show help
	@ grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[0;32m%-30s\033[0m %s\n", $$1, $$2}'
