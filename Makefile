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

.PHONY: all driver-registrar clean test
.IGNORE : buildEnvClean

REGISTRY_NAME=ihub.helium.io:29006/library
IMAGE_NAME_REGISTRAR=driver-registrar
IMAGE_NAME_ATTACHER=csi-attacher
IMAGE_VERSION=v1.0.0
IMAGE_NAME_TAG_ATTACHER=$(REGISTRY_NAME)/$(IMAGE_NAME_ATTACHER):$(IMAGE_VERSION)
IMAGE_NAME_TAG_REGISTRAR=$(REGISTRY_NAME)/$(IMAGE_NAME_REGISTRAR):$(IMAGE_VERSION)

ROOTPATH=$(shell pwd)
BUILDGOPATH=/tmp/k8splugin-build
BUILDPATH=$(BUILDGOPATH)/src/github.com/kubernetes-csi
  

all: build


buildEnvClean:
	@rm -rf $(BUILDPATH) 1>/dev/null 2>/dev/null || true

buildEnv: buildEnvClean
	@mkdir -p $(BUILDGOPATH)/src/github.com/kubernetes-csi
	@ln -s $(ROOTPATH)/driver-registrar $(BUILDGOPATH)/src/github.com/kubernetes-csi/driver-registrar
	@ln -s $(ROOTPATH)/external-attacher $(BUILDGOPATH)/src/github.com/kubernetes-csi/external-attacher
	
buildRegistrar: buildEnv
	@cd $(BUILDPATH)/driver-registrar && GOPATH=$(BUILDGOPATH) CGO_ENABLED=0 GOOS=linux go build -o ./bin/driver-registrar ./cmd/driver-registrar
	

buildRegistrarImage: buildRegistrar
	@rm Dockerfile 1>/dev/null 2>/dev/null || true
	@cp Dockerfile_registrar Dockerfile
	@docker build -t $(IMAGE_NAME_TAG_REGISTRAR) .

buildAttacher: buildEnv
	@cd $(BUILDPATH)/external-attacher && GOPATH=$(BUILDGOPATH) CGO_ENABLED=0 GOOS=linux go build -o ./bin/csi-attacher ./cmd/csi-attacher
	
buildAttacherImage: buildAttacher
	@rm Dockerfile
	@cp Dockerfile_attacher Dockerfile
	@docker build -t $(IMAGE_NAME_TAG_ATTACHER) .

push:
	docker push $(IMAGE_NAME_TAG_ATTACHER)
	docker push $(IMAGE_NAME_TAG_REGISTRAR)
	
release: container push
	
	

container: buildRegistrarImage buildAttacherImage
	
clean: buildEnvClean
	@rm Dockerfile 1>/dev/null 2>/dev/null || true
	@rm -rf ./driver-registrar/bin 1>/dev/null 2>/dev/null || true
	@rm -rf ./external-attacher/bin 1>/dev/null 2>/dev/null || true
  
