BINARY_NAME = faas-dispatch
IMAGE = lolopop/faas-dispatch
TAG = latest

publish: docker-build docker-push

docker-build:
	DOCKER_BUILDKIT=1 docker build -t ${IMAGE}:${TAG} .

docker-push:
	docker push ${IMAGE}:${TAG}

install:
	kubectl apply -f kubernetes/0
	kubectl apply -f kubernetes/1
	kubectl apply -f kubernetes/2
	kubectl apply -f kubernetes/3
	kubectl apply -f kubernetes/4
	kubectl apply -f kubernetes/5
uninstall: 
	kubectl delete -f kubernetes/0
	kubectl delete -f kubernetes/1
	kubectl delete -f kubernetes/2
	kubectl delete -f kubernetes/3
	kubectl delete -f kubernetes/4
	kubectl delete -f kubernetes/5
vendor:
	go mod vendor -v

build:
	go build -o bin/${BINARY_NAME}

fmt:
	gofmt -l -d $(shell find . -type f -name '*.go' -not -path "./vendor/*")

clean:
	rm -rf vendor/ bin/
