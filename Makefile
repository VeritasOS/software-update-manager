# Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.DEFAULT_GOAL := help
.PHONY: help
help:	## Display this help message.
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

TOP=$(CURDIR)
BUILDDATE ?= $(shell /usr/bin/date -u +%Y%m%d%H%M%S)

# Go build related variables
GOSRC=$(TOP)

# Set GOBIN to where binaries get picked up while creating RPM/ISO.
GOBIN?=$(TOP)/bin
GOCOVER=$(GOSRC)/cover
GOTOOLSBIN=$(TOP)/tools/go/

# SUM SDK related variables
SDK_NAME = sum-sdk
# Keep the SDK_VERSION same as ASUM_RPM_FORMAT_VERSION.
SDK_VERSION = 2.0
SDK_REVERSION = 1
SDK_SOURCE_PATH	= $(TOP)/sdk

.SILENT:

.PHONY: all
all: clean analyze build test

.PHONY: clean
clean: 	## Clean Software Update Manager go build & test artifacts
	@echo "Cleaning Software Update Manager Go binaries...";
	export GOBIN=$(GOBIN); \
	cd $(GOSRC); \
	go mod vendor; \
	go clean -i -mod=vendor ./...;
	@echo "Cleaning Go test artifacts... ";
	-@rm $(GOSRC)/{,.}*{dot,html,log,svg,xml};
	-@rm -rf $(GOSRC)/vendor/plugin-manager;
	-@rm -rf $(GOCOVER);

.PHONY: build
build: 	## Build source code
	# Since go build determines and build only updated sources, no need to run clean all go binaries
	@echo "Building Software Update Manager Go binaries...";
	export GOBIN=$(GOBIN); \
	cd $(GOSRC); \
	go mod vendor; \
	go install -ldflags "-X main.buildDate=`date -u +%Y%m%d.%H%M%S`" -mod=vendor -v ./...; \
	ret=$$?; \
	if [ $${ret} -ne 0 ]; then \
		@echo "Failed to build Software Update Manager Go binaries."; \
		exit 1; \
	fi


.PHONY: analyze
analyze: gofmt golint govet go-race  ## Analyze source code for different errors through gofmt, golint, govet, go-race

.PHONY: golint
golint:	## Run golint
	@echo Checking Software Update Manager Go code for lint errors...
	$(GOTOOLSBIN)/golint -set_exit_status `cd $(GOSRC); go list -mod=vendor  -f '{{.Dir}}' ./...`

.PHONY: gofmt
gofmt:	## Run gofmt
	@echo Checking Go code for format errors...
	fmterrs=`gofmt -l . | grep -v vendor/ 2>&1`; \
	if [ "$$fmterrs" ]; then \
		echo "gofmt must be run on the following files:"; \
		echo "$$fmterrs"; \
		exit 1; \
	fi

.PHONY: govet
govet:	## Run go vet
	@echo Vetting Software Update Manager Go code for errors...
	cd $(GOSRC); \
	go mod vendor; \
	go vet -mod=vendor -all ./...

.PHONY: test
test:  	## Run all tests
	echo "Running Software Update Manager Go Unit Tests...";
	mkdir -p $(GOCOVER);
	export INTEG_TEST_BIN=$(GOSRC); \
	export PM_CONF_FILE=$(GOSRC)/sample/pm.config.yaml; \
	export INTEGRATION_TEST=START; \
	cd $(GOSRC); \
	go mod vendor; \
	test_failed=0; \
	d=pm; \
	go test -mod=vendor -v --cover -covermode=count -coverprofile=$(GOCOVER)/$${d}.out ./... | \
		$(GOTOOLSBIN)/go-junit-report > TEST-$${d}.xml; \
	ret=$${PIPESTATUS[0]}; \
	if [ $${ret} -ne 0 ]; then \
		echo "Go unit test failed for $${d}."; \
		test_failed=1; \
	fi ; \
	awk -f $(TOP)/tools/gocoverage-collate.awk $(GOCOVER)/* > $(GOCOVER)/cover.out; \
	go tool cover -html=$(GOCOVER)/cover.out -o go-coverage-$${d}.html; \
	$(GOTOOLSBIN)/gocov convert $(GOCOVER)/cover.out | $(GOTOOLSBIN)/gocov-xml > go-coverage-$${d}.xml; \
	rm -rf $(GOCOVER)/*; \
	export INTEGRATION_TEST=DONE; \
	if [ $${test_failed} -ne 0 ]; then \
		echo "Go unit tests failed."; \
		exit 1; \
	fi

.PHONY: go-race
go-race: 	## Run Go tests with race detector enabled
	echo "Checking Go code for race conditions...";
	# NOTE: COVER directory should be present, along with INTEGRATION_TEST
	# 	value being set to "START" for integ_test.go to succeed.
	mkdir -p $(GOCOVER);
	export INTEGRATION_TEST=START; \
	export INTEG_TEST_BIN=$(GOSRC); \
	cd $(GOSRC); \
	go mod vendor; \
	export PM_CONF_FILE=$(GOSRC)/sample/pm.config.yaml; \
	go test -mod=vendor -v -race ./...;

.PHONY: .copy_update_binaries
.copy_update_binaries: build
	echo "Copying ASUM SDK related binaries...";
	cp -prf $(GOBIN)/* $(SDK_SOURCE_PATH)/scripts/; \
	if [ $$? -ne 0 ]; then \
 		echo "ERROR: $${f} failed to copy to $(SDK_SOURCE_PATH)/scripts/"; \
		exit 1; \
	fi;

.PHONY: ship-asum-sdk
ship-asum-sdk: .copy_update_binaries
	@echo "Creating ASUM SDK tar..."
	# cp $(TOP)/tools/mkrpm.sh $(SDK_SOURCE_PATH)/;
	ship_dir=$(TOP); \
	if [ -n "$(SHIP_DIR)" ]; then \
		ship_dir=$(SHIP_DIR); \
		mkdir -p $${ship_dir} || exit 1; \
	fi; \
	asum_sdk_tar="$${ship_dir}/$(SDK_NAME)-$(SDK_VERSION)-$(SDK_REVERSION)-$(BUILDDATE).tar.gz"; \
	/bin/tar -czf $${asum_sdk_tar} -C $(SDK_SOURCE_PATH) .; \
	echo "Successfully created ASUM SDK at $${asum_sdk_tar}.";

.NOTPARALLEL: