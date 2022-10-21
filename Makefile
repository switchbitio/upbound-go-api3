# ====================================================================================
# Setup Project

PROJECT_NAME := upbound-go-api3
PROJECT_REPO := github.com/upbound/$(PROJECT_NAME)

PLATFORMS ?= linux_amd64 linux_arm64
# -include will silently skip missing files, which allows us to load those files
# with a target in the Makefile. If only "include" was used, the make command
# would fail and refuse to run a target until the include commands succeeded.
-include build/makelib/common.mk

# ====================================================================================
# Setup Output

# S3_BUCKET ?= upbound.releases/$(PROJECT_NAME)
# -include build/makelib/output.mk

# ====================================================================================
# Setup Go

# Set a sane default so that the nprocs calculation below is less noisy on the initial
# loading of this file
NPROCS ?= 1

# each of our test suites starts a kube-apiserver and running many test suites in
# parallel can lead to high CPU utilization. by default we reduce the parallelism
# to half the number of CPU cores.
GO_TEST_PARALLEL := $(shell echo $$(( $(NPROCS) / 2 )))

# GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/$(PROJECT_NAME)
GO_STATIC_PACKAGES = $(GO_PROJECT)/cmd/service
GO_LDFLAGS += -X $(GO_PROJECT)/internal/version.version=$(VERSION)
GO_SUBDIRS += cmd internal
GO111MODULE = on
-include build/makelib/golang.mk

# ====================================================================================
# Setup Additional Tools

# ====================================================================================
# Setup Kubernetes tools

USE_HELM3 = true
HELM3_VERSION = v3.6.3
-include build/makelib/k8s_tools.mk

# ====================================================================================
# Setup Helm

HELM_BASE_URL = https://helm.upbound.io
HELM_S3_BUCKET = upbound.charts
# HELM_CHARTS = $(PROJECT_NAME)
HELM_CHARTS = service
HELM_CHART_LINT_ARGS_crossplane = --set nameOverride='',imagePullSecrets=''
-include build/makelib/helm.mk

# ====================================================================================
# Setup Images
# Due to the way that the shared build logic works, images should
# all be in folders at the same level (no additional levels of nesting).

REGISTRY_ORGS = docker.io/donovanmuller
# IMAGES = $(PROJECT_NAME)
IMAGES = service
-include build/makelib/imagelight.mk

# ====================================================================================
# Targets

# run `make help` to see the targets and options

# We want submodules to be set up the first time `make` is run.
# We manage the build/ folder and its Makefiles as a submodule.
# The first time `make` is run, the includes of build/*.mk files will
# all fail, and this target will be run. The next time, the default as defined
# by the includes will be run instead.
fallthrough: submodules
	@echo Initial setup complete. Running make again . . .
	@make

# Update the submodules, such as the common build scripts.
submodules:
	@git submodule sync
	@git submodule update --init --recursive

# NOTE(hasheddan): the build submodule currently overrides XDG_CACHE_HOME in
# order to force the Helm 3 to use the .work/helm directory. This causes Go on
# Linux machines to use that directory as the build cache as well. We should
# adjust this behavior in the build submodule because it is also causing Linux
# users to duplicate their build cache, but for now we just make it easier to
# identify its location in CI so that we cache between builds.
go.cachedir:
	@go env GOCACHE

.PHONY: submodules fallthrough gen-schema

dev:
	@PLATFORM=${PLATFORM} HOSTARCH=${HOSTARCH} VERSION=${VERSION} IMAGE_TEMP_DIR=${IMAGE_TEMP_DIR} OUTPUT_DIR=${OUTPUT_DIR} skaffold dev
