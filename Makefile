# 🌱 miku: Tiny, stateless microservice to notify that your Discord bot is going under maintenance, made in Go
# Copyright (c) 2022 Nino Team
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

VERSION    :=$(shell cat ./version.json | jq .version | tr -d '"')
GIT_COMMIT :=$(shell git rev-parse --short=8 HEAD)
BUILD_DATE := $(shell go run ./cmd/build-date.go)
GIT_TAG    ?= $(shell git describe --tags --match "v[0-9]*")

GOOS   := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

ifeq ($(GOOS), linux)
	TARGET_OS ?= linux
else ifeq ($(GOOS),darwin)
	TARGET_OS ?= darwin
else ifeq ($(GOOS),windows)
	TARGET_OS ?= windows
else
	$(error System $(GOOS) is not supported at this time)
endif

EXTENSION :=
ifeq ($(TARGET_OS),windows)
	EXTENSION := .exe
endif

# Usage: `make deps`
deps:
	@echo Updating dependency tree...
	go mod tidy
	go mod download
	@echo Updated dependency tree successfully.

# Usage: `make build`
build:
	@echo Now building Tsubaki for platform $(GOOS)/$(GOARCH)!
	go build -ldflags "-s -w -X main.version=${VERSION} -X main.commitSha=${GIT_COMMIT} -X \"main.buildDate=${BUILD_DATE}\"" -o ./bin/miku$(EXTENSION)
	@echo Successfully built the binary. Use './bin/miku$(EXTENSION)' to run!

# Usage: `make clean`
clean:
	@echo Now cleaning project..
	rm -rf bin/
	go clean
	@echo Done!

# Usage: `make fmt`
fmt:
	@echo Formatting project...
	go fmt
