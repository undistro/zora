# Docker
TAG ?= latest
IMG ?= operator:${TAG}
WORKER_IMG ?= worker:${TAG}
DOCKERFILE ?= cmd/manager/Dockerfile

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
ADDLICENSE ?= $(shell pwd)/bin/addlicense

# Tests
# Kubebuilder assets required by <envtest>.
ENVTEST ?= $(shell pwd)/bin/setup-envtest
ENVTEST_K8S_VERSION = 1.23

# Execution
PLUGINS ?= popeye
