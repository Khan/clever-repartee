# Injecting introspective version information at build time

It is a best practice to inject introspective version information as
variables into your Golang binary at build-time.
This is most useful for tagging your binary with a human readable version and
Git shasum.

This allows an executable binary artifact to be able to introspectively report
it's provenance, e.g. which exact version of which source code generated it.

If a kubernetes Job silently fails to deploy, and the older version is still
being executed, this can cause a great deal of confusion. If the job can
immediately log what source code generated it as it starts up, this will
save quite a bit of time.

```
export MODVER="v0.0.0-$(git --no-pager show\
 --quiet\
 --abbrev=12\
 --date='format-local:%Y%m%d%H%M%S'\
 --format="%cd-%h")"
export GIT_COMMIT="$(git rev-parse HEAD)"
export COMMIT_SHA=$(shell git rev-parse --short HEAD)
export BUILD=$(shell date +%FT%T%z)
export PROJECT=khan-internal-services
export APP=districts-clever-roster
go build -trimpath \
      -ldflags="-w -s \
      -X github.com/Khan/districts-jobs/pkg/version.AppName=${APP} \
      -X github.com/Khan/districts-jobs/pkg/version.Date=${BUILD_DATE} \
      -X github.com/Khan/districts-jobs/pkg/version.GitCommit=${COMMIT_SHA} \
      -X github.com/Khan/districts-jobs/pkg/version.Project=${PROJECT} \
      -X github.com/Khan/districts-jobs/pkg/version.Version=${VERSION} \
      -extldflags '-static'" -a \
      -o /go/bin/main ./services/districts/cmd/clever/main.go
```

In AppEngine, this is handled quite differently, but in Docker in GKE, we must
handle it more manually at build time.

