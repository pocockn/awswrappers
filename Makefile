BRANCH = "master"
VERSION = $(shell cat ./VERSION)
PATH_BASE ?= "/go/src/github.com/vidsy"
GO_BUILDER_IMAGE ?= "vidsyhq/go-builder"
REPONAME ?= "awswrappers"
TEST_PACKAGES ?= "./sqs ./dynamodb ./s3 ./sns"

DEFAULT: test

check-version:
	git fetch && (! git rev-list ${VERSION})

push-tag:
	git checkout ${BRANCH}
	git pull origin ${BRANCH}
	git tag ${VERSION}
	git push origin ${BRANCH} --tags

build:
	@docker run \
	-v "${CURDIR}":${PATH_BASE}/${REPONAME} \
	-e BUILD=false \
	-w ${PATH_BASE}/${REPONAME} \
	${GO_BUILDER_IMAGE}

test:
	@go test "${TEST_PACKAGES}"

test-docker:
	@docker run \
	-it \
	--rm \
	-v "${CURDIR}":${PATH_BASE}/${REPONAME} \
	-w ${PATH_BASE}/${REPONAME} \
	--entrypoint=go \
	${GO_BUILDER_IMAGE} test "${TEST_PACKAGES}"

test_ci:
	@docker run \
	-v "${CURDIR}":${PATH_BASE}/${REPONAME} \
	-w ${PATH_BASE}/${REPONAME} \
	--entrypoint=go \
	${GO_BUILDER_IMAGE} test "${TEST_PACKAGES}" -cover
