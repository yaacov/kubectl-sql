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

VERSION_GIT := $(shell git describe --tags)
VERSION ?= ${VERSION_GIT}

kubesql_cmd := $(wildcard ./cmd/kubectl-sql/*.go)
kubesql_pkg := $(wildcard ./pkg/**/*.go)

all: kubectl-sql

kubectl-sql: $(kubesql_cmd) $(kubesql_pkg)
	go build -ldflags='-X github.com/yaacov/kubectl-sql/pkg/cmd.clientVersion=${VERSION}' -o kubectl-sql $(kubesql_cmd)

.PHONY: lint
lint:
	golint ./pkg/...
	golint ./cmd/...

.PHONY: fmt
fmt:
	gofmt -s -w $(kubesql_cmd) $(kubesql_pkg)

.PHONY: dist
dist: kubectl-sql
	tar -zcvf kubectl-sql.tar.gz LICENSE kubectl-sql
	sha256sum kubectl-sql.tar.gz > kubectl-sql.tar.gz.sha256sum

.PHONY: clean
clean:
	rm -f kubectl-sql
	rm -f kubectl-sql.tar.gz
	rm -f kubectl-sql.tar.gz.sha256sum

