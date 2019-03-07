all: build

TAG?=v1.0.0
REGISTRY?=ihub.helium.io:29006
NAME?=library/hostpathcsi
FLAGS=
ENVVAR=
GOOS?=linux
ROOTPATH=`pwd` 
BUILDGOPATH=/tmp/k8splugin-build
BUILDPATH=$(BUILDGOPATH)/src/github.com/k8s-plugins/csi-plugin/hostpathpv

.IGNORE : deletedeploy
.IGNORE : createnamespace
.IGNORE : buildEnvClean

buildEnvClean:
	@rm $(BUILDPATH) 1>/dev/null 2>/dev/null || true

buildEnv: buildEnvClean
	@mkdir -p $(BUILDGOPATH)/src/github.com/k8s-plugins/csi-plugin/
	@ln -s $(ROOTPATH) $(BUILDPATH)

build: clean deps buildEnv
	@cd $(BUILDPATH) && GOPATH=$(BUILDGOPATH) $(ENVVAR) GOOS=$(GOOS) CGO_ENABLED=0 go build -o hostpathcsi hostpath/main.go

docker:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
	docker build --pull -t ${REGISTRY}/${NAME}:${TAG} .
	docker push ${REGISTRY}/${NAME}:${TAG}

createnamespace:
	@kubectl create namespace k8splugin 1>/dev/null 2>/dev/null || true
	@kubectl annotate ns k8splugin io.enndata.namespace/alpha-allowhostpath=true 1>/dev/null 2>/dev/null || true
	@kubectl annotate ns k8splugin io.enndata.namespace/alpha-allowprivilege=true 1>/dev/null 2>/dev/null || true

deletedeploy:
	@kubectl delete -f deploy/hostpath-csi-attacher.yaml 1>/dev/null 2>/dev/null || true
	@kubectl delete -f deploy/hostpath-csi-plugin.yaml 1>/dev/null 2>/dev/null || true

install: createnamespace deletedeploy
	@cat deploy/hostpath-csi-attacher.yaml | sed "s/ihub.helium.io:29006/$(REGISTRY)/g" > deploy/tmp-attacher.yaml
	@cat deploy/hostpath-csi-plugin.yaml | sed "s/ihub.helium.io:29006/$(REGISTRY)/g" > deploy/tmp-plugin.yaml
	kubectl create -f deploy/tmp-attacher.yaml
	kubectl create -f deploy/tmp-plugin.yaml
	@rm deploy/tmp-attacher.yaml deploy/tmp-plugin.yaml

uninstall: deletedeploy 

release: build docker
	rm -f hostpathcsi

clean:
	rm -f hostpathcsi

keeptruetest:
	./test/csikeeptest.sh
	
format:
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -d {} + | tee /dev/stderr)" || \
	test -z "$$(find . -path ./vendor -prune -type f -o -name '*.go' -exec gofmt -s -w {} + | tee /dev/stderr)"

.PHONY: all deps build test-unit clean format release

