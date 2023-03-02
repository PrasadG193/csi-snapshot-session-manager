IMAGE_REPO_SERVER ?= prasadg193/cbt-server
IMAGE_REPO_CLIENT ?= prasadg193/cbt-client
IMAGE_REPO_AGGAPI ?= prasadg193/cbt-datapath
IMAGE_TAG_SERVER ?= latest
IMAGE_TAG_CLIENT ?= latest
IMAGE_TAG_AGGAPI ?= latest

API_GROUP ?= cbt
API_VERSION ?= v1alpha1
API_KIND ?= VolumeSnapshotToken

GOOS ?= linux
GOARCH ?= amd64

image:
	docker build -t $(IMAGE_REPO_AGGAPI):$(IMAGE_TAG_AGGAPI) -f Dockerfile .
	#docker build -t $(IMAGE_REPO_SERVER):$(IMAGE_TAG_SERVER) -f Dockerfile-server .
	#docker build -t $(IMAGE_REPO_CLIENT):$(IMAGE_TAG_CLIENT) -f Dockerfile-client .

push:
	docker push $(IMAGE_REPO_AGGAPI):$(IMAGE_TAG_AGGAPI)
	#docker push $(IMAGE_REPO_SERVER):$(IMAGE_TAG_SERVER)
	#docker push $(IMAGE_REPO_CLIENT):$(IMAGE_TAG_CLIENT)

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -o cbt-server ./cmd/server/main.go
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -a -o cbt-client ./cmd/client/main.go

codegen:
	./hack/update-codegen.sh

codegen-verify:
	./hack/verify-codegen.sh

init_repo:
	apiserver-boot init repo --domain storage.k8s.io

create_group:
	apiserver-boot create group version resource --group $(API_GROUP) --version $(API_VERSION) --kind $(API_KIND)

.PHONY: yaml
yaml:
	rm -rf yaml-generated
	apiserver-boot build config --name cbt-aggapi --namespace cbt-svc --image $(IMAGE_REPO_AGGAPI):$(IMAGE_TAG_AGGAPI) --output yaml-generated

