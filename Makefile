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
#   - go 1.16 or higher
#   - curl or wget
#   - CGO enabled
#   - musl-gcc package installed for static binary compilation
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
	@echo "Building for ${GOOS}/${GOARCH}"
	go build -ldflags='-X github.com/yaacov/kubectl-sql/pkg/cmd.clientVersion=${VERSION}' -o kubectl-sql $(kubesql_cmd)

kubectl-sql-static: $(kubesql_cmd) $(kubesql_pkg)
	CGO_ENABLED=1 CC=musl-gcc go build \
		-tags netgo \
		-ldflags '-linkmode external -extldflags "-static" -X github.com/yaacov/kubectl-sql/pkg/cmd.clientVersion=${VERSION}' \
		-o kubectl-sql \
		$(kubesql_cmd)

.PHONY: lint
lint:
	go vet ./pkg/... ./cmd/...
	golangci-lint run ./pkg/... ./cmd/...

.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: dist
dist: kubectl-sql
	tar -zcvf kubectl-sql.tar.gz LICENSE kubectl-sql
	sha256sum kubectl-sql.tar.gz > kubectl-sql.tar.gz.sha256sum

.PHONY: clean
clean:
	rm -f kubectl-sql
	rm -f kubectl-sql.tar.gz
	rm -f kubectl-sql.tar.gz.sha256sum

.PHONY: test
test:
	go test -v -cover ./pkg/... ./cmd/...
	go test -coverprofile=coverage.out ./pkg/... ./cmd/...
	go tool cover -func=coverage.out
	@rm coverage.out

