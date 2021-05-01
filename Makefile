# Copyright (c) 2021 Veritas Technologies LLC. All rights reserved. IP63-2828-7171-04-15-9

# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.DEFAULT_GOAL := help
.PHONY: help
help:	## Display this help message.
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

TOP=$(CURDIR)
BUILDDATE ?= $(shell /usr/bin/date -u +%Y%m%d%H%M%S)
TAR = /bin/tar

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

TAR       = /bin/tar

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
	echo "Copying SUM SDK related binaries...";
	cp -prf $(GOBIN)/* $(SDK_SOURCE_PATH)/scripts/; \
	if [ $$? -ne 0 ]; then \
 		echo "ERROR: $${f} failed to copy to $(SDK_SOURCE_PATH)/scripts/"; \
		exit 1; \
	fi;

SUM_PATH = $(TOP)/
SAMPLE_UPDATE= $(SUM_PATH)/sample/update/
PRODUCT_VERSION = 1.0
tmp_SUM_SDK = $(TOP)/tmp/sum-sdk
tmp_SHIP_DIR = $(TOP)/tmp/ship

# Ship SDK and create sample update RPM
SUM_PATH = $(TOP)/
SAMPLE_UPDATE = $(SUM_PATH)/sample/update
PRODUCT_VERSION = 1.0
tmp_SUM_SDK = $(TOP)/tmp/sum-sdk
tmp_SHIP_DIR = $(TOP)/tmp/ship

.PHONY: ship-sum-sdk
ship-sum-sdk: .copy_update_binaries
	@echo "Creating SUM SDK tar..."
	# cp $(TOP)/tools/mkrpm.sh $(SDK_SOURCE_PATH)/;
	ship_dir=$(TOP); \
	if [ -n "$(SHIP_DIR)" ]; then \
		ship_dir=$(SHIP_DIR); \
		mkdir -p $${ship_dir} || exit 1; \
	fi; \
	sum_sdk_tar="$${ship_dir}/$(SDK_NAME)-$(SDK_VERSION)-$(SDK_REVERSION)-$(BUILDDATE).tar.gz"; \
	/bin/tar -czf $${sum_sdk_tar} -C $(SDK_SOURCE_PATH) .; \
	echo "Successfully created SUM SDK at $${sum_sdk_tar}.";

.PHONY: getsumsdk
getsumsdk:
	# NOTE: Usually the SUM SDK will be extracted from an already built tar.
	#       But here, in this case, we need to first build the
	#       SUM SDK tar and then consume it here.
	echo "=============== Get SUM SDK ===============";
	ship_dir=$(tmp_SHIP_DIR); \
	rm -rf $${ship_dir}; \
	$(MAKE) -C $(SUM_PATH) ship-sum-sdk SHIP_DIR=$${ship_dir}; \
	echo "===============  extracting sum-sdk tar file =========================="; \
	mkdir -p $(tmp_SUM_SDK); \
	$(TAR) -xzvf $${ship_dir}/sum-sdk-*.tar.gz -C $(tmp_SUM_SDK); \
	if [ $$? -ne 0 ]; then \
		echo "ERROR: Failed to extract sum-sdk to $(tmp_SUM_SDK)"; \
		exit 1; \
	fi;

# TODO: Update below target to download from github.
.PHONY: get_remote_sum_sdk
get_remote_sum_sdk:
	# Avoid downloading it again when it's already present. Run `get_latest_asum_sdk` to get latest.
ifeq ($(shell ls $(tmp_SUM_SDK)/.no-latest 2> /dev/null),) 
	$(ECHO) "===== Downloading ASUM RPM Generation utility (asum-sdk) $(tmp_SUM_SDK) =====";
	$(RM) -rf $(tmp_SUM_SDK); 
	$(MKDIR) -p $(tmp_SUM_SDK);
	px_api_url="$(HTTP_ARTIFACTORY)/api/storage/release/platformx/main"; \
	px_ver=$$($(CURL) $${px_api_url} | jq -c '.children[] | select(.folder)' | sort -V | tail -2 | head -1 | jq .uri); \
	px_ver=$${px_ver:2:-1}; \
	px_buildtag=$${px_ver##*-}; \
	$(ECHO) "PlatformX latest version-buildtag: $${px_ver}"; \
	rpm_name=$$($(CURL) $(HTTP_ARTIFACTORY)/api/storage/release/platformx/main/$${px_ver}/ \
		| jq '.children[].uri' | grep asum-sdk); \
	rpm_name=$${rpm_name%\"}; \
	rpm_name=$${rpm_name#\"}; \
	$(ECHO) "RPM: $${rpm_name}"; \
	$(ECHO) "===== Downloading $${rpm_name} ====="; \
	asum_url="$(HTTP_ARTIFACTORY)/release/platformx/main/$${px_ver}/"; \
	$(WGET) $${asum_url}/$${rpm_name} -P $(tmp_SUM_SDK); \
	if [ $$? -ne 0 ] ; then \
		$(ECHO) "ERROR: $(WGET) $${rpm_name} failed to download."; \
		exit 1; \
	fi ; \
	$(ECHO) Successfully downloaded $${rpm_name} into $(tmp_SUM_SDK);

	echo "===============  Extracting asum-sdk =========================="; \
	$(TAR) -xzvf $(tmp_SUM_SDK)/asum-sdk-*.tar.gz -C $(tmp_SUM_SDK); \
	if [ $$? -ne 0 ]; then \
	    echo "ERROR: Failed to extract asum-sdk to $(tmp_SUM_SDK)"; \
		exit 1; \
	fi;
	$(ECHO) "To avoid downloading latest SUM SDK again and again on local " \
		"builds, you can touch $(tmp_SUM_SDK)/.no-latest file. "
	# touch $(tmp_SUM_SDK)/.no-latest
endif

.PHONY: sampleupdate_usinglocalsdk
sampleupdate_usinglocalsdk: getsumsdk
	echo "===============  Creating a sample update RPM   ===============";
	git checkout -- $(TOP)/sample/update/library/version/version.install $(TOP)/sample/update/rpm-info.json
	myVersion=$(PRODUCT_VERSION).9; \
	sed -i -e "s%__VERSION__%$${myVersion}%g" $(TOP)/sample/update/library/version/version.install; \
	sed -i -e "s%__PRODUCT_VERSION__%$(PRODUCT_VERSION)%g" $(TOP)/sample/update/rpm-info.json; \
	JenkinsOptions=""; \
	ship_dir=$(tmp_SHIP_DIR); \
	if [ -n "$(JENKINS_URL)" ]; then \
		JenkinsOptions="BUILDDATE=$${BUILDTAG##*-}"; \
		ship_dir=$(JENKINS_UPLOAD_DEST); \
	fi; \
	$(MAKE) -C $(SAMPLE_UPDATE) update \
		SUM_SDK_PATH=$(tmp_SUM_SDK) \
		SHIP_DIR=$${ship_dir} \
		UPDATE_VERSION=$${myVersion} \
		$${JenkinsOptions};
	git checkout -- $(TOP)/sample/update/library/version/version.install $(TOP)/sample/update/rpm-info.json


.PHONY: sampleupdates
sampleupdates: sampleupdate_usinglocalsdk sampleupdate sampleupdate_reboot_commit sampleupdate_no_version_change

.PHONY: sampleupdate
sampleupdate: get_remote_sum_sdk
	$(ECHO) "======== Generating sample update RPM using SUM SDK..."; \
	git checkout -- samples/update/library/node-version/node-version.install
	myVersion=$(VERSION).0.9; \
	sed -i -e "s%__VERSION__%$${myVersion}%g" $(TOP)/samples/update/library/node-version/node-version.install; \
	ship_dir=$(TOP); \
	if [ -n "$(JENKINS_URL)" ]; then \
		Options="RPM_RELEASE=$${BUILDTAG##*-}"; \
		ship_dir=$(TOP)/$(PRODUCT)/$(BRANCH)/$(BUILDTAG); \
	fi; \
	$(MAKE) -C $(tmp_SUM_SDK) generate \
		PLUGINS_LIBRARY=$(TOP)/samples/update/library/ \
		RPM_INFO_FILE=$(TOP)/samples/update/rpm-info.json \
		RPM_URL=https://github.com/VeritasOS/software-update-manager \
		RPM_VERSION=$${myVersion} \
		SHIP_DIR=$${ship_dir} \
		$${Options} || exit $$?
	git checkout -- samples/update/library/node-version/node-version.install

.PHONY: sampleupdate_reboot_commit
sampleupdate_reboot_commit: get_remote_sum_sdk
	$(ECHO) "======== Generating sample update RPM requiring reboot using SUM SDK...";
	git checkout -- samples/update-reboot-commit/library/node-version/node-version.install
	myVersion=$(VERSION).9; \
	sed -i -e "s%__VERSION__%$${myVersion}%g" $(TOP)/samples/update-reboot-commit/library/node-version/node-version.install; \
	ship_dir=$(TOP); \
	if [ -n "$(JENKINS_URL)" ]; then \
		Options="RPM_RELEASE=$${BUILDTAG##*-}"; \
		ship_dir=$(TOP)/$(PRODUCT)/$(BRANCH)/$(BUILDTAG); \
	fi; \
	$(MAKE) -C $(tmp_SUM_SDK) generate \
		PLUGINS_LIBRARY=$(TOP)/samples/update-reboot-commit/library/ \
		RPM_INFO_FILE=$(TOP)/samples/update-reboot-commit/rpm-info.json \
		RPM_URL=https://github.com/VeritasOS/software-update-manager \
		RPM_VERSION=$${myVersion} \
		SHIP_DIR=$${ship_dir} \
		$${Options} || exit $$?
	git checkout -- samples/update-reboot-commit/library/node-version/node-version.install

.PHONY:sampleupdate_no_version_change
sampleupdate_no_version_change:
	$(ECHO) "======== Generating sample update RPM using SUM SDK..."; \
	myVersion=2.0.0.0.$${BUILDTAG##*-}; \
	ship_dir=$(tmp_SHIP_DIR); \
	if [ -n "$(JENKINS_URL)" ]; then \
		Options="RPM_RELEASE=$${BUILDTAG##*-}"; \
		ship_dir=$(TOP)/$(PRODUCT)/$(BRANCH)/$(BUILDTAG); \
	fi; \
	$(MAKE) -C $(tmp_SUM_SDK) generate \
		PLUGINS_LIBRARY=$(TOP)/samples/update-no-version-change/library/ \
		RPM_INFO_FILE=$(TOP)/samples/update-no-version-change/rpm-info.json \
		RPM_URL=https://github.com/VeritasOS/software-update-manager \
		RPM_VERSION=$${myVersion} \
		SHIP_DIR=$${ship_dir} \
		$${Options} || exit $$?

.PHONY: clean-update
clean-update:
	$(RM) -rf $(tmp_SUM_SDK)

.NOTPARALLEL:
