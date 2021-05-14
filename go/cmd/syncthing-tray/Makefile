# © 2021 Ilya Mateyko. All rights reserved.
# Use of this source code is governed by the MIT
# license that can be found in the LICENSE.md file.

GOOS    = "windows"
GOARCH  = "amd64"
LDFLAGS = "-s -w -H windowsgui -buildid="

.PHONY: build clean help

build: ## Build
	@ GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -trimpath -ldflags=$(LDFLAGS)

clean: ## Clean
	@ rm *.exe
