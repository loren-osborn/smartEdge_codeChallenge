# I prefer not to have duplicate constants in different files, but for
# exercise purposes this is the quickest way to move forward.
CONTAINER_GOPATH       := $(shell docker run --rm golang:1.10 sh -c 'echo $$GOPATH')
IMAGE_TAG               = codechal
PROJECT_URI             = github.com/smartedge/codechallenge
CONTAINER_PROJECT_DIR   = $(CONTAINER_GOPATH)/src/$(PROJECT_URI)
ALL_SOURCE_FILES       := $(shell find . -type f "(" -name "*.go" -or -name "*.iml" -or -path "*/testdata/*" ")" | sed -e 's,^\.\/,,')
PROD_SOURCE_FILES       = $(filter-out %_test.go,$(filter %.go, $(ALL_SOURCE_FILES)))
TEST_SOURCE_FILES       = $(filter %_test.go,$(ALL_SOURCE_FILES)) $(filter-out %.go, $(ALL_SOURCE_FILES))
PROD_SOURCE_FILES       = $(filter-out %_test.go,$(filter %.go, $(ALL_SOURCE_FILES)))
GENERATED_FILES         = event_timestamps coverage.out coverage.html
# Not including "production_container_image", as it is a final output, and may be in use
# elsewhere on the system:
DISCARDABLE_IMAGE_TAGS  = golang_base_image tester_image
PRECIOUS_IMAGE_TAGS     = production_container_image
RM_IMAGES_IF_PRESENT    = @bash -c 'for img_name in $(1) ; do if docker image inspect $$img_name >/dev/null 2>/dev/null ; then >&2 echo docker image rm $$img_name ; docker image rm $$img_name ; fi ; done'


default: event_timestamps/production_container_image

.PRECIOUS: coverage.html

.SECONDARY: test

.PHONY: clean purge

event_timestamps/production_container_image: test event_timestamps Dockerfile $(PROD_SOURCE_FILES)
	docker build -t production_container_image .
	touch $@

event_timestamps/golang_base_image: Dockerfile event_timestamps
	docker build --target golang_base -t golang_base_image .
	touch $@

# source files are pulled in as a volume, so the tester_image isn't dependant on them
event_timestamps/tester_image: event_timestamps/golang_base_image Dockerfile.test event_timestamps
	docker build -f Dockerfile.test -t tester_image .
	touch $@

test coverage.out coverage.html: event_timestamps/tester_image $(ALL_SOURCE_FILES)
	docker run --rm --tty -v "$$(pwd):$(CONTAINER_PROJECT_DIR)" tester_image:latest

event_timestamps:
	mkdir -p event_timestamps

clean:
	rm -rf $(GENERATED_FILES)
	$(call RM_IMAGES_IF_PRESENT,$(DISCARDABLE_IMAGE_TAGS))

# This is the "dangerous" version of "make clean"
purge: clean
	$(call RM_IMAGES_IF_PRESENT,$(PRECIOUS_IMAGE_TAGS))
