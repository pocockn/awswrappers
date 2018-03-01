BRANCH = "master"
VERSION = $(shell cat ./VERSION)
PATH_BASE ?= "/go/src/github.com/vidsy"
GO_BUILDER_IMAGE ?= "vidsyhq/go-builder"
REPONAME ?= "awswrappers"
PACKAGES ?= "$(shell glide nv)"

DEFAULT: test

install:
	@dep ensure -v

check-version:
	git fetch && (! git rev-list ${VERSION})

push-tag:
	git checkout ${BRANCH}
	git pull origin ${BRANCH}
	git tag ${VERSION}
	git push origin ${VERSION}

test:
	@go test "${PACKAGES}"

vet:
	@go vet "${PACKAGES}"

