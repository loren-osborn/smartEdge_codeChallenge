# I prefer not to have duplicate constants in different files, but for
# exercise purposes this is the quickest way to move forward.
CONTAINER_GOPATH      := $(shell docker run --rm golang:1.10 sh -c 'echo $$GOPATH')
IMAGE_TAG             := codechal
PROJECT_URI           := github.com/smartedge/codechallenge
CONTAINER_PROJECT_DIR := $(CONTAINER_GOPATH)/src/$(PROJECT_URI)

default: production_container_image

production_container_image: test
	docker build -t production_container_image .

golang_base_image: Dockerfile
	docker build --target golang_base -t golang_base_image .

tester_image: golang_base_image Dockerfile.test
	docker build -f Dockerfile.test -t tester_image .

test: tester_image
	docker run --rm --tty -v "$$(pwd):$(CONTAINER_PROJECT_DIR)" tester_image
