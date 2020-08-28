############################
# STEP 1 build executable binary
############################
# golang debian buster 1.13.6 linux/amd64
# https://github.com/docker-library/golang/blob/master/1.13/buster/Dockerfile
FROM golang:1.14.4-buster as builder
# FROM golang@sha256:f6cefbdd25f9a66ec7dcef1ee5deb417882b9db9629a724af8a332fe54e3f7b3

# Ensure ca-certficates are up to date
RUN update-ca-certificates

ENV GOFLAGS="-mod=readonly"
ENV GOPRIVATE="github.com/Khan"
ENV GO111MODULE=on

ARG APPDIR=clever-repartee
ENV APPDIR=${APPDIR}

# Moving outside of $GOPATH forces modules on without having to set ENVs
WORKDIR /src/${APPDIR}

# Add go.mod and go.sum first to maximize caching
COPY ./${APPDIR}/go.mod ./${APPDIR}/go.sum /src/${APPDIR}/

RUN go mod download -x
RUN go mod verify

# The temp .dockerignore in buildcontext dir excludes non-go files (and secrets)
COPY . /src
# expect build-time variables, but set some defaults
ARG APP=clever-repartee
ARG BUILD_DATE
ARG COMMIT_SHA
ARG VERSION

# use the arg values to set the ENV var default
ENV APP=${APP} \
    BUILD_DATE=${BUILD_DATE} \
    COMMIT_SHA=${COMMIT_SHA} \
    PROJECT=${PROJECT} \
    VERSION=${VERSION}

# Build the static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
      -ldflags="-w -s \
      -X github.com/Khan/clever-repartee/pkg/version.AppName=${APP} \
      -X github.com/Khan/clever-repartee/pkg/version.Date=${BUILD_DATE} \
      -X github.com/Khan/clever-repartee/pkg/version.GitCommit=${COMMIT_SHA} \
      -X github.com/Khan/clever-repartee/pkg/version.Project=${PROJECT} \
      -X github.com/Khan/clever-repartee/pkg/version.Version=${VERSION} \
      -extldflags '-static'" -a \
      -o /go/bin/main ./main.go

############################
# STEP 2 build a small image
############################
# user:group is nobody:nobody, uid:gid = 65534:65534
FROM gcr.io/distroless/static:nonroot
# FROM gcr.io/distroless/static@sha256:08322afd57db6c2fd7a4264bf0edd9913176835585493144ee9ffe0c8b576a76
WORKDIR /go/bin
# Copy our static executable
COPY --from=builder /go/bin/main .

# Document environment variables we expect, but default to empty
ENV CLEVER_ID=${CLEVER_ID}
ENV CLEVER_SECRET=${CLEVER_SECRET}
ENV MAP_CLEVER_ID=${MAP_CLEVER_ID}
ENV MAP_CLEVER_SECRET=${MAP_CLEVER_SECRET}
ENV PORT=8080
# Document port exposure
EXPOSE ${PORT}
# Run completely unprivileged.
USER nobody:nobody
# Run the hello binary.
ENTRYPOINT ["/go/bin/main"]
