#
# Copyright 2020 Yaacov Zamir <kobi.zamir@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# Prerequisites:
#   - go 1.23 or higher
#   - curl or wget
#
# Run `make install-tools` to install required development tools

VERSION_GIT := $(shell git describe --tags)
VERSION ?= ${VERSION_GIT}

all: kubectl-sql

.PHONY: install-tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

kubesql_cmd := $(wildcard ./cmd/kubectl-sql/*.go)
kubesql_pkg := $(wildcard ./pkg/**/*.go)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

kubectl-sql: clean $(kubesql_cmd) $(kubesql_pkg)
	@echo "Building static binary for ${GOOS}/${GOARCH}"
	CGO_ENABLED=0 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-sql/pkg/cmd.clientVersion=${VERSION}' \
		-o kubectl-sql \
		$(kubesql_cmd)

# Cross-compilation targets
.PHONY: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64
build-linux-amd64: clean $(kubesql_cmd) $(kubesql_pkg)
	@echo "Building for linux/amd64"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-sql/pkg/cmd.clientVersion=${VERSION}' \
		-o kubectl-sql-linux-amd64 \
		$(kubesql_cmd)

build-linux-arm64: clean $(kubesql_cmd) $(kubesql_pkg)
	@echo "Building for linux/arm64"
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-sql/pkg/cmd.clientVersion=${VERSION}' \
		-o kubectl-sql-linux-arm64 \
		$(kubesql_cmd)

build-darwin-amd64: clean $(kubesql_cmd) $(kubesql_pkg)
	@echo "Building for darwin/amd64"
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-sql/pkg/cmd.clientVersion=${VERSION}' \
		-o kubectl-sql-darwin-amd64 \
		$(kubesql_cmd)

build-darwin-arm64: clean $(kubesql_cmd) $(kubesql_pkg)
	@echo "Building for darwin/arm64"
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-sql/pkg/cmd.clientVersion=${VERSION}' \
		-o kubectl-sql-darwin-arm64 \
		$(kubesql_cmd)

build-windows-amd64: clean $(kubesql_cmd) $(kubesql_pkg)
	@echo "Building for windows/amd64"
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-a \
		-ldflags '-s -w -X github.com/yaacov/kubectl-sql/pkg/cmd.clientVersion=${VERSION}' \
		-o kubectl-sql-windows-amd64.exe \
		$(kubesql_cmd)

# Build all platforms
.PHONY: build-all
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

# Create release archives for all platforms
.PHONY: dist-all
dist-all: build-all
	@echo "Creating release archives..."
	tar -zcvf kubectl-sql-${VERSION}-linux-amd64.tar.gz LICENSE kubectl-sql-linux-amd64
	tar -zcvf kubectl-sql-${VERSION}-linux-arm64.tar.gz LICENSE kubectl-sql-linux-arm64
	tar -zcvf kubectl-sql-${VERSION}-darwin-amd64.tar.gz LICENSE kubectl-sql-darwin-amd64
	tar -zcvf kubectl-sql-${VERSION}-darwin-arm64.tar.gz LICENSE kubectl-sql-darwin-arm64
	zip kubectl-sql-${VERSION}-windows-amd64.zip LICENSE kubectl-sql-windows-amd64.exe
	@echo "Generating individual checksums..."
	sha256sum kubectl-sql-${VERSION}-linux-amd64.tar.gz > kubectl-sql-${VERSION}-linux-amd64.tar.gz.sha256sum
	sha256sum kubectl-sql-${VERSION}-linux-arm64.tar.gz > kubectl-sql-${VERSION}-linux-arm64.tar.gz.sha256sum
	sha256sum kubectl-sql-${VERSION}-darwin-amd64.tar.gz > kubectl-sql-${VERSION}-darwin-amd64.tar.gz.sha256sum
	sha256sum kubectl-sql-${VERSION}-darwin-arm64.tar.gz > kubectl-sql-${VERSION}-darwin-arm64.tar.gz.sha256sum
	sha256sum kubectl-sql-${VERSION}-windows-amd64.zip > kubectl-sql-${VERSION}-windows-amd64.zip.sha256sum

.PHONY: lint
lint:
	go vet ./pkg/... ./cmd/...
	$(shell go env GOPATH)/bin/golangci-lint run ./pkg/... ./cmd/...

.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: dist
dist: kubectl-sql
	tar -zcvf kubectl-sql.tar.gz LICENSE kubectl-sql
	sha256sum kubectl-sql.tar.gz > kubectl-sql.tar.gz.sha256sum

.PHONY: clean
clean:
	rm -f kubectl-sql kubectl-sql-*
	rm -f *.tar.gz *.zip *.sha256sum

.PHONY: test
test:
	go test -v -cover ./pkg/... ./cmd/...
	go test -coverprofile=coverage.out ./pkg/... ./cmd/...
	go tool cover -func=coverage.out
	@rm coverage.out

