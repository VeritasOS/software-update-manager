# Copyright (c) 2021 Veritas Technologies LLC. All rights reserved IP63-2828-7171-04-15-9.

# ASUM_RPM_FORMAT_VERSION is like a protocol version.
# 	This will be used by the ASUM binary to list the contents appropriately.
# 	This version needs to be updated, only if there is format/layout changes
# 	that breaks backward compatibility w.r.t. parsing RPM info by the `asum`
# 	binary.
ASUM_RPM_FORMAT_VERSION=2

# Jenking build options
PRODUCT?=asum
VERSION?=1.0
BRANCH?=main

TOP	= $(shell pwd)
PACKAGE_DIR=$(TOP)/tmp/
SHIP_DIR?=$(PACKAGE_DIR)

# Reference: https://rpm-packaging-guide.github.io/#rpm-packaging-workspace
RPM_NAME?=VRTS$(PRODUCT)-update
RPM_VERSION?=$(VERSION).1
RPM_RELEASE?=2
RPM_URL?=https://www.veritas.com/support/en_US.html
RPM_SUMMARY?=$(PRODUCT) $(RPM_VERSION) Update for $(PRODUCT) $(VERSION).
RPM_INFO_FILE?=$(TOP)/rpm-info.json

RPM_SCRIPTS_DIR?=$(TOP)/scripts

# The plugins' library directory is used for RPM packaging.
# Copy plugins here before invoking the copy_plugins target.
PLUGINS_LIBRARY?=$(PACKAGE_DIR)/library
PLUGIN_TYPES?="preinstall install prereboot postreboot rollback-precheck prerollback rollback commit-precheck commit"

RPM_PRE_SCRIPT?=$(RPM_SCRIPTS_DIR)/rpm_pre.sh
RPM_SPEC_FILE=asumrpm.spec
RPM_INST_FILE=asumrpm.inst

TARGET_DIR=system/upgrade/repository/update/$(RPM_NAME)-$(RPM_VERSION)-$(RPM_RELEASE)

RPM2EMBED?=""

.PHONY : all
all: generate


# generate: target adds contents into an RPM through mkrpm utility.
# 	The update requires two directories scripts and library containing plugins
# 	at the top level. The plugins' library should have all the required plugins
# 	(i.e., .preinstall, .install, and so on plugins) that drive the update workflow
# 	required for updating the system. 
.PHONY: generate
generate:
	@echo ===== Generating ASUM RPM =====
	@echo Creating spec and inst files.
	if [ ! -f $(RPM_INFO_FILE) ] ; then \
		echo "RPM Info JSON file does not exist at $(RPM_INFO_FILE)." ; \
		exit 1; \
	fi
	
	errno=0 ; \
	/usr/bin/cat $(RPM_INFO_FILE) | jq empty || errno=$$? ; \
	if [ $${errno} -ne 0 ] ; then \
		echo "Invalid JSON format found in Version Info JSON file at $(RPM_INFO_FILE)" ; \
		exit $${errno} ; \
	fi ; \
	rpm_info_json=$$(/usr/bin/cat $(RPM_INFO_FILE) | jq -c '.') ; \
	if [[ $${rpm_info_json} == "" ]] ; then \
		echo "RPM Info JSON is empty. Make sure a valid JSON is present in file at $(RPM_INFO_FILE)" ; \
		exit 1 ; \
	fi ; \
	rpm_desc="$(subst ",\",$${rpm_info_json})" ; \
	mkdir -p $(PACKAGE_DIR)/{$(TARGET_DIR),SPECS} ; \
	cp $(TOP)/$(RPM_SPEC_FILE) $(PACKAGE_DIR)/SPECS ; \
	sed -i -e "s%__SUMMARY__%$(RPM_SUMMARY)%g" \
		-e "s%__NAME__%$(RPM_NAME)%g" \
		-e "s%__VERSION__%$(RPM_VERSION)%g" \
		-e "s|__RELEASE__|$(RPM_RELEASE)|g" \
		-e "s%__URL__%$(RPM_URL)%g" \
		-e "s%__ASUM_RPM_FORMAT_VERSION__%$(ASUM_RPM_FORMAT_VERSION)%g" \
		-e "s%__RPM_INFO__%$${rpm_desc}%g" \
		$(PACKAGE_DIR)/SPECS/$(RPM_SPEC_FILE);

	# Add RPM pre script if it exists or empty string to the PRE section of the SPEC file
	sed -e '/__PRESCRIPT__/ {' -e 'r $(RPM_PRE_SCRIPT)' -e 'd' -e '}' -i $(PACKAGE_DIR)/SPECS/$(RPM_SPEC_FILE)

	cp $(TOP)/$(RPM_INST_FILE) $(PACKAGE_DIR)/SPECS
	sed -i -e "s%__TARGETDIR__%$(TARGET_DIR)%g" $(PACKAGE_DIR)/SPECS/$(RPM_INST_FILE)

	# If the ISO path is available, extract and store contents.
	iso_target_dir=iso/contents ; \
	if [ -f "$(ISO_PATH)" ] ; then \
		echo ============ Extracting $(ISO_PATH) in $(PLUGINS_LIBRARY)/$${iso_target_dir} ============= ; \
		mkdir -p $(PLUGINS_LIBRARY)/$${iso_target_dir}; \
		rm -rf $(PLUGINS_LIBRARY)/$${iso_target_dir}/* ; \
		7z x $(ISO_PATH) -o$(PLUGINS_LIBRARY)/$${iso_target_dir}/ -y; \
	fi ; \

	@echo ======== Adding ASUM scripts and plugins for version $(RPM_VERSION) for RPM packaging.
	cp -v ${RPM_SCRIPTS_DIR}/* $(PACKAGE_DIR)/$(TARGET_DIR)
	cp -Rv $(PLUGINS_LIBRARY) $(PACKAGE_DIR)/$(TARGET_DIR)

	@echo ======== Generating plugins dependencies graph.
	rm -rf $(PLUGINS_LIBRARY)/../imgs/;
	mkdir -p $(PLUGINS_LIBRARY)/../imgs/;
	pluginTypes=$(PLUGIN_TYPES); \
	for pt in $${pluginTypes}; do \
		echo "Generating image for $${pt}..."; \
		$(TOP)/scripts/asum pm list -library $(PLUGINS_LIBRARY) -type $${pt} -log-dir $(PLUGINS_LIBRARY)/../imgs/ -log-file $${pt}; \
	done;

	@echo ======== Keeping only plugin graph images and cleaning logs in image dir.
	rm -f $(PLUGINS_LIBRARY)/../imgs/*.{dot,log};
	curYear=$$(/usr/bin/date -u +%Y); \
	for i in $$(ls $(PLUGINS_LIBRARY)/../imgs/*); do \
		if [ -s $${i} ]; then \
			mv $${i} $${i%.$${curYear}*}.svg; \
		else \
			rm $${i}; \
		fi; \
	done;

	@echo ======== Generating/Updating README.md.
	readme=$(PLUGINS_LIBRARY)/../README.md; \
	rm -f $${readme}; \
	cp template.md $${readme}; \
	pluginTypes=$(PLUGIN_TYPES); \
	for pt in $${pluginTypes}; do \
		grep "./imgs/$${pt}.svg" $${readme}; \
		if [ $$? != 0 ]; then \
			echo "" >> $${readme}; \
			echo "## $${pt}" >> $${readme}; \
			echo "" >> $${readme}; \
			echo "![$${pt}](./imgs/$${pt}.svg)" >> $${readme}; \
		fi; \
	done; \
	git add $${readme} $(PLUGINS_LIBRARY)/../imgs/*.svg;

	@echo ======== Calling mkrpm.sh utility to generate the ASUM RPM.
	cd $(PACKAGE_DIR); \
	if [ -n "$(SHIP_DIR)" ]; then \
		export PLATFORM_RPM_LOCATION=$(SHIP_DIR); \
	fi; \
	$(TOP)/mkrpm.sh -m $(PACKAGE_DIR)/SPECS/$(RPM_SPEC_FILE)


.PHONY: clean
clean: 
	@echo ======== Removing ASUM RPM artifacts.
	-rm -rf $(PACKAGE_DIR)


.PHONY: strip_watermark
strip_watermark:
	echo "===============  Stripping watermark from source code  ==============="; \
	find . \( ! -path "*.git*" -a ! -ipath './Makefile' \) -type f -print0  | xargs -0 -I@ sed -i -e 's|IP63-2828-7171-04-15-9||I' @


.NOTPARALLEL:
