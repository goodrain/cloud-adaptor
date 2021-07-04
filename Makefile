build:
	GOOS=linux CGO_ENABLED=1 GOARCH=amd64 go build -o bin/cloudadaptor-x86.linux
	GOOS=darwin GOARCH=amd64 go build -o bin/cloudadaptor-x86.darwin
release: build
	ossutil cp -r -u bin/ oss://grstatic/binary

swag-init:
	swag init -g cmd/cloud-adaptor/main.go --parseDependency --parseDepth 1

dev:
	HELM_REPO_FILE=./data/helm/repo/repository.yaml HELM_CACHE=./data/helm/cache go run ./cmd/cloud-adaptor