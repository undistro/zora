# Docker
REG_PORT ?= 5000
REG_HOST ?= localhost
REG_ADDR ?= ${REG_HOST}:${REG_PORT}
MINIK_ADDR ?= 192.168.49.2
IMG_TAG ?= latest
IMG ?= ${REG_ADDR}/operator:${IMG_TAG}
WORKER_IMG ?= ${REG_ADDR}/worker:${IMG_TAG}
DOCKERFILE ?= Dockerfile

# Commands
# Requirement for 'setup-envtest.sh' in the test target.
SHELL = /bin/bash -o pipefail
.SHELLFLAGS = -ec
IGNORE_NOT_FOUND ?= true

ifeq (,$(shell go env GOBIN))
 GOBIN=$(shell go env GOPATH)/bin
else
 GOBIN=$(shell go env GOBIN)
endif

# Code Generation
PROJECT_ROOT := $(shell git rev-parse --show-toplevel)
PROJECT_PACKAGE ?= $(shell go list -m)
CONTROLLER_GEN ?= $(shell pwd)/bin/controller-gen
KUSTOMIZE ?= $(shell pwd)/bin/kustomize

# Tests
# Kubebuilder assets required by <envtest>.
ENVTEST ?= $(shell pwd)/bin/setup-envtest
ENVTEST_K8S_VERSION = 1.23

# Execution
PLUGINS ?= popeye
