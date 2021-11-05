# Copyright 2021 VMware. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

REQUIRED_BINARIES := imgpkg kbld ytt vendir
PACKAGE_VERSION := $(VERSION)
OCI_IMAGE := ghcr.io/$(IMAGE_TAG_BASE)-package

check-carvel:
	$(foreach exec,$(REQUIRED_BINARIES),\
		$(if $(shell which $(exec)),,$(error "'$(exec)' not found. Carvel toolset is required. See instructions at https://carvel.dev/#install")))

sync: check-carvel # Performs a `vendir sync` for each package version.
	$(foreach pkgv,$(PACKAGE_VERSION),cd $(pkgv)/bundle && vendir sync)

lock: check-carvel # Updates the image lock file for each package.
	$(foreach pkgv,$(PACKAGE_VERSION),cd $(pkgv) && kbld --file bundle --imgpkg-lock-output bundle/.imgpkg/images.yml)

.PHONY: push
push: check-carvel # Build and push packages.	
	$(foreach pkgv,$(PACKAGE_VERSION),cd $(pkgv) && imgpkg push --bundle $(OCI_IMAGE):${pkgv} --file bundle/)
