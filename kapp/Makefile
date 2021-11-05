# Copyright 2021 VMware. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

REQUIRED_BINARIES := imgpkg kbld ytt vendir
PACKAGE_VERSION := $(VERSION)
OCI_IMAGE := ghcr.io/$(IMAGE_TAG_BASE)-package

CARVEL_TOOLS_DIR=/tmp/carvel_tools

check-carvel: install-carvel
	$(foreach exec,$(REQUIRED_BINARIES),\
		$(if $(shell ls $(CARVEL_TOOLS_DIR)/$(exec)),,$(error "'$(exec)' not found in $(CARVEL_TOOLS_DIR). Install it...")))

install-carvel:
ifeq (,$(shell ls $(CARVEL_TOOLS_DIR)/ytt 2>/dev/null))
	@{ \
	mkdir -p $(CARVEL_TOOLS_DIR) ;\
	wget -O- https://carvel.dev/install.sh -O /tmp/install.sh && K14SIO_INSTALL_BIN_DIR=$(CARVEL_TOOLS_DIR) bash /tmp/install.sh ;\
	}	
endif
	

#sync: check-carvel # Performs a `vendir sync` for each package version.
#	$(foreach pkgv,$(PACKAGE_VERSION),cd $(pkgv)/bundle && $(CARVEL_TOOLS_DIR)/vendir sync)

lock: check-carvel # Updates the image lock file for each package.
	$(CARVEL_TOOLS_DIR)/kbld --file bundle --imgpkg-lock-output bundle/.imgpkg/images.yml

push: lock # Build and push packages.	
	echo "push $(OCI_IMAGE):${PACKAGE_VERSION}"
	$(CARVEL_TOOLS_DIR)/imgpkg push --bundle $(OCI_IMAGE):${PACKAGE_VERSION} --file bundle/
