# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: all csi-attacher clean test
.IGNORE : buildEnvClean

REGISTRY_NAME=ihub.helium.io:29006/library
IMAGE_NAME=csi-attacher
IMAGE_VERSION=v0.3.0
IMAGE_TAG=$(REGISTRY_NAME)/$(IMAGE_NAME):$(IMAGE_VERSION)

ROOTPATH=`pwd` 
BUILDGOPATH=/tmp/k8splugin-build
BUILDPATH=$(BUILDGOPATH)/src/github.com/kubernetes-csi/external-attacher

REV=$(shell git describe --long --match='v*' --tags --dirty)

ifdef V
TESTARGS = -v -args -alsologtostderr -v 5
else
TESTARGS =
endif


all: build

buildEnvClean:
	@rm $(BUILDPATH) 1>/dev/null 2>/dev/null || true

buildEnv: buildEnvClean
	@mkdir -p $(BUILDGOPATH)/src/github.com/kubernetes-csi/
	@ln -s $(ROOTPATH) $(BUILDGOPATH)/src/github.com/kubernetes-csi/external-attacher
	
	
build: buildEnv
	mkdir -p bin
	@cd $(BUILDPATH) && GOPATH=$(BUILDGOPATH) CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-X main.version=$(REV) -extldflags "-static"' -o ./bin/csi-attacher ./cmd/csi-attacher

clean:
	-rm -rf bin
	
release: container push

container: build
	docker build -t $(IMAGE_TAG) .

push: container
	docker push $(IMAGE_TAG)

test:
	go test `go list ./... | grep -v 'vendor'` $(TESTARGS)
	go vet `go list ./... | grep -v vendor`
