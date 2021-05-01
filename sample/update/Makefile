# Copyright (c) 2020 Veritas Technologies LLC. All rights reserved IP63-2828-7171-04-15-9.

# Jenking build options
PRODUCT ?= VRTS
VERSION ?= 2.0
BRANCH ?= main

TOP	= $(shell pwd)
SUM_SDK_PATH ?= $(TOP)/../../sdk/

UPDATE_VERSION ?= $(VERSION).1
UPDATE_RELEASE ?= $(shell /usr/bin/date -u +%Y%m%d%H%M%S)
UPDATE_URL ?= https://www.veritas.com/support/en_US.html
UPDATE_RPM_NAME = $(PRODUCT)-update
UPDATE_RPM_INFO_FILE = $(TOP)/rpm-info.json

.PHONY: all
all: update


.PHONY: update
update:
	@echo "======== Generating update RPM using ASUM SDK..."; \
	$(MAKE) -C $(SUM_SDK_PATH) generate \
		PLUGINS_LIBRARY=$(TOP)/library/ \
		RPM_NAME=$(UPDATE_RPM_NAME) \
		RPM_VERSION=$(UPDATE_VERSION) \
		RPM_RELEASE=$(UPDATE_RELEASE) \
		RPM_URL=$(UPDATE_URL) \
		RPM_SUMMARY="Sample Update RPM created using ASUM SDK" \
		RPM_INFO_FILE=$(UPDATE_RPM_INFO_FILE) || exit $$?


.PHONY: clean
clean:
	$(MAKE) -C $(SUM_SDK_PATH) clean


.PHONY: strip_watermark
strip_watermark:
	echo "===============  Stripping watermark from source code  ==============="; \
	find . \( ! -path "*.git*" -a ! -ipath './Makefile' \) -type f -print0  | xargs -0 -I@ sed -i -e 's|IP63-2828-7171-04-15-9||I' @


.NOTPARALLEL:
