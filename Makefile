# I prefer not to have duplicate constants in different files, but for
# exercise purposes this is the quickest way to move forward.
CONTAINER_GOPATH       := $(shell docker run --rm golang:1.10 sh -c 'echo $$GOPATH')
IMAGE_TAG               = codechal
PROJECT_URI             = github.com/smartedge/codechallenge
CONTAINER_PROJECT_DIR   = $(CONTAINER_GOPATH)/src/$(PROJECT_URI)
ALL_SOURCE_FILES       := $(shell   \
	find . -type f "("              \
		-name "*.go"           -or  \
		-name "*.iml"          -or  \
		-name "*.toml"         -or  \
		-path "*/buildtools/*" -or  \
		-path "*/testdata/*"        \
	")" | sed -e 's,^\.\/,,')
PROD_SOURCE_FILES       = $(filter-out %_test.go,$(filter %.go, $(ALL_SOURCE_FILES)))
TEST_SOURCE_FILES       = $(filter %_test.go,$(ALL_SOURCE_FILES)) $(filter-out %.go, $(ALL_SOURCE_FILES))
PROD_SOURCE_FILES       = $(filter-out %_test.go,$(filter %.go, $(ALL_SOURCE_FILES)))
GENERATED_FILES         = event_timestamps coverage.out coverage.html godoc .smartEdge
# Putting "production_container_image" in seperate PRECIOUS_IMAGE_TAGS, as it
# is a final output, and may be in use elsewhere on the system:
DISCARDABLE_IMAGE_TAGS  = golang_base_image tester_image demo_image
PRECIOUS_IMAGE_TAGS     = production_container_image
# This is a GNU make $(call ) function:
RM_IMAGES_IF_PRESENT    = @bash -c 'for img_name in $(1) ; do if docker image inspect $$img_name >/dev/null 2>/dev/null ; then >&2 echo docker image rm $$img_name ; docker image rm $$img_name ; fi ; done'


default: event_timestamps/production_container_image demo

.PRECIOUS: coverage.html

.SECONDARY: test gh-pages

.PHONY: clean purge

event_timestamps/production_container_image: event_timestamps Makefile \
	Dockerfile $(PROD_SOURCE_FILES) test
	docker build -t production_container_image .
	touch $@

event_timestamps/golang_base_image: Makefile Dockerfile event_timestamps
	docker build --target golang_base -t golang_base_image .
	touch $@

event_timestamps/demo_image: Makefile Dockerfile.demo event_timestamps/golang_base_image \
	event_timestamps $(PROD_SOURCE_FILES)
	docker build -f Dockerfile.demo -t demo_image .
	touch $@

# Source files are pulled in as a volume, so the tester_image isn't dependant
# on them
event_timestamps/tester_image: event_timestamps/golang_base_image Makefile \
	Dockerfile.test event_timestamps
	docker build -f Dockerfile.test -t tester_image .
	touch $@

test coverage.out coverage.html godoc: event_timestamps/tester_image \
	Makefile $(ALL_SOURCE_FILES)
	$(strip docker run                            \
		--rm                                      \
		--tty                                     \
		-v "$$(pwd):$(CONTAINER_PROJECT_DIR)"     \
		--mount type=tmpfs,destination=/tmp/tmpfs \
		--env TERM="$$TERM"                       \
		--env EXT_UID_GID="$$(id -u):$$(id -g)"   \
		tester_image:latest)

event_timestamps/gh-pages: coverage.html godoc Makefile
	@git diff --quiet master -- . || (echo "Branch gh-pages need to track remote master branch." ; exit 1)
	$(strip bash -xc \
		"git clone . ../tmp_smcc_ghpages                                            && \
		pushd ../tmp_smcc_ghpages                                                   && \
		git checkout gh-pages                                                       && \
		git rm -rq README.md coverage.html godoc                                    && \
		popd                                                                        && \
		cp -R README.md coverage.html godoc ../tmp_smcc_ghpages                     && \
		pushd ../tmp_smcc_ghpages                                                   && \
		git add README.md coverage.html godoc                                       && \
		exit_status=\"\$$?\"                                                        ;  \
		if [ -z \"\$$exit_status\" ] || [ \"0\" -eq \"\$$exit_status\" ] ; then        \
			if git diff --quiet HEAD -- . ; then                                       \
				popd                                                                ;  \
				rm -rf ../tmp_smcc_ghpages                                          ;  \
				exit \$$?                                                           ;  \
			fi                                                                      ;  \
			git commit -m \"Automated update\"                                      && \
			popd                                                                    && \
			git remote add tmp_ghpages ../tmp_smcc_ghpages                          && \
			git fetch tmp_ghpages                                                   && \
			git branch -f gh-pages tmp_ghpages/gh-pages                             ;  \
			exit_status=\"\$$?\"                                                    ;  \
		fi                                                                          ;  \
		popd                                                                        ;  \
		git remote rm tmp_ghpages                                                   ;  \
		exit_status_2=\"\$$?\"                                                      ;  \
		rm -rf ../tmp_smcc_ghpages                                                  ;  \
		exit_status_3=\"\$$?\"                                                      ;  \
		if [ -n \"\$$exit_status\" ] && [ \"0\" -ne \"\$$exit_status\" ] ; then        \
			exit \$$exit_status                                                     ;  \
		fi                                                                          ;  \
		if [ -n \"\$$exit_status_2\" ] && [ \"0\" -ne \"\$$exit_status_2\" ] ; then    \
			exit \$$exit_status_2                                                   ;  \
		fi                                                                          ;  \
		exit \$$exit_status_3" )
	touch $@

gh-pages: event_timestamps/gh-pages

demo: event_timestamps/demo_image
	$(strip docker run                            \
		--rm                                      \
		-i --tty                                  \
		demo_image:latest )
	@echo

build_local: event_timestamps Makefile $(PROD_SOURCE_FILES) test
	@sh -c \
		'if [ "$$(pwd)" != "$$GOPATH/src/$(PROJECT_URI)"  ] ; then \
			>&2 echo "To build this project locally, you must check it out" \
			"in $$GOPATH/src/$(PROJECT_URI) based on your current GOPATH" ;\
			exit 1 ; \
		fi'
	go build -o codechallenge $(PROJECT_URI)/cmd/codechallenge

event_timestamps: Makefile
	@$(strip bash -c \
		'if [ -d event_timestamps ] ; then       \
			>&2 echo touch event_timestamps ;    \
			         touch event_timestamps ;    \
		else                                     \
			>&2 echo mkdir -p event_timestamps ; \
			         mkdir -p event_timestamps ; \
		fi')
	@$(strip bash -c \
		'if ! [ -f .git/hooks/pre-push ] ; then                   \
			>&2 echo cp buildtools/pre-push .git/hooks/pre-push ; \
			         cp buildtools/pre-push .git/hooks/pre-push ; \
		fi')

clean:
	rm -rf $(GENERATED_FILES)
	$(call RM_IMAGES_IF_PRESENT,$(DISCARDABLE_IMAGE_TAGS))

# This is the "dangerous" version of "make clean"
purge: clean
	$(call RM_IMAGES_IF_PRESENT,$(PRECIOUS_IMAGE_TAGS))
