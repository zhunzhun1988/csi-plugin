all: build
 
GOOS?=linux
ROOTPATH=`pwd` 
BUILDGOPATH=/tmp/k8splugin-build
BUILDPATH=$(BUILDGOPATH)/src/k8s-plugins/csi-plugin/hostpathpvchangenode

.IGNORE : buildEnvClean


buildEnvClean:
	@rm $(BUILDGOPATH)/src/k8s-plugins/csi-plugin/hostpathpvchangenode 1>/dev/null 2>/dev/null || true

buildEnv: buildEnvClean
	@mkdir -p $(BUILDGOPATH)/src/k8s-plugins/csi-plugin/
	@ln -s $(ROOTPATH) $(BUILDGOPATH)/src/k8s-plugins/csi-plugin/hostpathpvchangenode

deps:
	go get github.com/tools/godep

build: buildEnv clean deps
	@cd $(BUILDPATH) && GOPATH=$(BUILDGOPATH) $(ENVVAR) GOOS=$(GOOS) CGO_ENABLED=0 godep go build ./...
	@cd $(BUILDPATH) && GOPATH=$(BUILDGOPATH) $(ENVVAR) GOOS=$(GOOS) CGO_ENABLED=0 godep go build -o hostpathpvchangenode
  
clean: buildEnvClean
	@rm -f hostpathpvchangenode

format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

.PHONY: all deps build  clean format buildEnvClean buildEnv

