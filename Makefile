SHELL := sh
.ONESHELL:
.EXPORT_ALL_VARIABLES:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules


# Note: Lazy Set - values within it are recursively expanded when the variable is used
# not when it's declared. Important because git sha can change.
GO111MODULE=on
COMMIT_SHA?=$(shell git rev-parse --short HEAD)
VERSION=v0.0.0
BUILD_DATE?=$(shell date +'%s')
LDFLAGS=-X main.version=${VERSION} -X main.commit=${COMMIT_SHA} -X main.date=${BUILD_DATE}
# Repository -
#   A collection of tags grouped under a common prefix (the name component before :).
#   For example, in an image tagged with the name my-app:3.1.4, my-app is the Repository component of the name.
#   A repository name is made up of slash-separated name components, optionally prefixed by the service's DNS hostname.
#   The hostname must follow comply with standard DNS rules, but may not contain _ characters.
#   If a hostname is present, it may optionally be followed by a port number in the format :8080.
#   Name components may contain lowercase characters, digits, and separators.
#   A separator is defined as a period, one or two underscores, or one or more dashes.
#   A name component may not start or end with a separator.
#
#
# Tag -
#   A tag serves to map a descriptive, user-given name to any single image ID.
#
# Image Name -
#   Informally, the name component after any prefixing hostnames and namespaces.
# 
# WARNING: A docker tag name must be valid ASCII and may contain lowercase and uppercase letters,
# digits, underscores, periods and dashes.
# A docker tag name may not start with a period or a dash and may contain a maximum of 128 characters.

PROJECT=khan-internal-services
APP=clever-repartee
IMAGE_NAME=${APP}
REPOSITORY_NAMESPACE=${PROJECT}
REGISTRY=gcr.io
REPOSITORY=${REGISTRY}/${REPOSITORY_NAMESPACE}/${IMAGE_NAME}
GOPATH?=${HOME}/go
GOPRIVATE?=github.com/Khan
INSTALLPATH=${GOPATH}/bin/${APP}
CLEVER_ID?=${CLEVER_ID}
CLEVER_SECRET?=${CLEVER_SECRET}
MAP_CLEVER_ID?=${MAP_CLEVER_ID}
MAP_CLEVER_SECRET?=${MAP_CLEVER_SECRET}

.PHONY: clean
clean: setup ## - Cleans go files and the binary
	@printf "\033[32m\xE2\x9c\x93 Cleaning your code\n\033[0m"
	mkdir -p ${GOPATH}/bin
	rm -f ${INSTALLPATH} || true
	@PATH="${GOPATH}/bin:${PATH}" GOPRIVATE=$(GOPRIVATE) gofmt -l -w -s . || true
	@PATH="${GOPATH}/bin:${PATH}" GOPRIVATE=$(GOPRIVATE) goimports -l -w . || true
	@PATH="${GOPATH}/bin:${PATH}" GOPRIVATE=$(GOPRIVATE) golines -m 80 --shorten-comments -w . || true
	GOPRIVATE=$(GOPRIVATE) go mod tidy || true

.PHONY: setup
setup: ## - Install any missing prerequisites like oapi-codegen goimports jq go
	@printf "\033[32m\xE2\x9c\x93 Installing necessary prerequisites if missing\n\033[0m"
	@mkdir -p $$HOME/go/bin
ifeq (, $(shell which jq))
	brew install jq
endif
ifeq (, $(shell which go))
	brew install go
endif
ifeq (, $(shell which oapi-codegen))
	go get -u github.com/deepmap/oapi-codegen/cmd/oapi-codegen
endif
ifeq (, $(shell which goimports))
	go get -u golang.org/x/tools/cmd/goimports
endif
ifeq (, $(shell which golines))
	go get -u github.com/segmentio/golines
endif
ifeq (, $(shell which golangci-lint))
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.30.0
endif

.PHONY: regenerate
regenerate: ## - Regenerate clever client and type files from openapi3 spec, protobuf stuff
	@printf "\033[32m\xE2\x9c\x93 Regenerate clever client and type files\n\033[0m"
	go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --generate types,client --package=generated -o ./pkg/generated/clever.gen.go full-v2.oas3.yml

.PHONY: build
build: clean ## - Build the application
	@printf "\033[32m\xE2\x9c\x93 Building your code\n\033[0m"
	GOPRIVATE=$(GOPRIVATE) go build -trimpath -ldflags='-w -s \
	-X github.com/Khan/clever-repartee/pkg/version.AppName=${APP} \
	-X github.com/Khan/clever-repartee/pkg/version.Date=${BUILD_DATE} \
	-X github.com/Khan/clever-repartee/pkg/version.GitCommit=${COMMIT_SHA} \
	-X github.com/Khan/clever-repartee/pkg/version.Project=${PROJECT} \
	-X github.com/Khan/clever-repartee/pkg/version.Version=${VERSION} \
	-extldflags "-static"' -a \
	-o ${INSTALLPATH} ./main.go

.PHONY: run
run: setup git-pull secrets build ## - Runs go run main.go
	@printf "\033[32m\xE2\x9c\x93 Running your code\n\033[0m"
	@PATH="${GOPATH}/bin:${PATH}" CLEVER_ID=$(CLEVER_ID) CLEVER_SECRET=$(CLEVER_SECRET) MAP_CLEVER_ID=$(MAP_CLEVER_ID) MAP_CLEVER_SECRET=$(MAP_CLEVER_SECRET) ${APP} diff -district=$$DISTRICT_ID -json -force

.PHONY: justrun
justrun: ## - Just Runs go run main.go
	@printf "\033[32m\xE2\x9c\x93 Running your code\n\033[0m"
	@PATH="${GOPATH}/bin:${PATH}" CLEVER_ID=$(CLEVER_ID) CLEVER_SECRET=$(CLEVER_SECRET) MAP_CLEVER_ID=$(MAP_CLEVER_ID) MAP_CLEVER_SECRET=$(MAP_CLEVER_SECRET) ${APP} diff -district=$$DISTRICT_ID -json -force

.PHONY: test
test: ## - Runs go test with default values
	@printf "\033[32m\xE2\x9c\x93 Testing your code to find potential problems\n\033[0m"
	GOPRIVATE=$(GOPRIVATE) go test -v -count=1 -race ./...

.PHONY: diff-test-district
diff-test-district: setup git-pull secrets build ## - Rosters test district as gcs protobuf file and json file
	@printf "\033[32m\xE2\x9c\x93 Rostering test district\n\033[0m"
	@PATH="${GOPATH}/bin:${PATH}" CLEVER_ID=$(CLEVER_ID) CLEVER_SECRET=$(CLEVER_SECRET) MAP_CLEVER_ID=$(MAP_CLEVER_ID) MAP_CLEVER_SECRET=$(MAP_CLEVER_SECRET) ${APP} diff -district=$$DISTRICT_ID -json -force

.PHONY: cover
cover: test ## - Runs test coverage report
	@printf "\033[32m\xE2\x9c\x93 Running Code Test Coverage Report\n\033[0m"
	go test -count=1 -coverprofile=coverage.out
	GOPRIVATE=$(GOPRIVATE) go tool cover -html=coverage.out

.PHONY: lint
lint: clean ## - Lint the application code for problems and nits
	@printf "\033[32m\xE2\x9c\x93 Linting your code to find potential problems\n\033[0m"
	GOPRIVATE=$(GOPRIVATE) go vet ./...
	@PATH="${GOPATH}/bin:${PATH}" golangci-lint run --fix

.PHONY: docker-build
docker-build:	## - Build the smallest secure golang docker image based on distroless static
	@printf "\033[32m\xE2\x9c\x93 Build the smallest and secured golang docker image based on distroless static\n\033[0m"
	docker build -f ./Dockerfile -t ${REGISTRY}/${APP}:${COMMIT_SHA} ..

.PHONY: docker-build-no-cache
docker-build-no-cache:	## - Build the smallest secure golang docker image based on distroless static with no cache
	@printf "\033[32m\xE2\x9c\x93 Build the smallest and secured golang docker image based on scratch\n\033[0m"
	docker build --no-cache -f Dockerfile -t ${REGISTRY}/${APP}:${COMMIT_SHA} .

.PHONY: ls
ls: ## - List size docker images
	@printf "\033[32m\xE2\x9c\x93 Look at the size dude !\n\033[0m"
	@echo image ls ${REGISTRY}/${APP}
	docker image ls ${REGISTRY}/${APP}

.PHONY: docker-run
docker-run:	docker-build ## - Run the docker image based on distroless static nonroot
	@printf "\033[32m\xE2\x9c\x93 Run the docker image based on distroless static nonroot\n\033[0m"
	docker run --entrypoint "/go/bin/main" \
	--user 65534:65534 \
	--publish 127.0.0.1:8080:8080/tcp \
	"${REPOSITORY}:${COMMIT_SHA}" version

.PHONY: docker-push
docker-push: docker-build ## - Pushes the docker image to registry
	docker push "${REPOSITORY}:${COMMIT_SHA}"

.PHONY: git-pull
git-pull: ## - Git Pull Makes everything minty fresh
	git pull

.PHONY: help
## help: Prints this help message
help: ## - Show help message
	@printf "\033[32m\xE2\x9c\x93 usage: make [target]\n\n\033[0m"
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
