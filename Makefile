appname := postisto

sources = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
artifact_version = $(shell cat VERSION | tr -d '\n')
timestamp = $(shell date +%Y%m%d-%H%M%S)
gitrev = $(shell git rev-parse --short HEAD)
build_version = $(artifact_version)-$(timestamp)-$(gitrev)

build = GOOS=$(1) GOARCH=$(2) GOARM=$(4) go build -trimpath -ldflags "-X=main.build=$(build_version)" -o build/$(appname)$(3) cmd/$(appname)/main.go
tar = cd build && tar -cvzf $(appname)-$(artifact_version).$(1)-$(2).tar.gz $(appname)$(3) && rm $(appname)$(3)
zip = cd build && zip $(appname)-$(artifact_version).$(1)-$(2).zip $(appname)$(3) && rm $(appname)$(3)

all: build test install

.PHONY: build test go.test clean fmt go.mod vendor-update vendor docker-build release github-release docker-release install

build: clean docker-build windows darwin linux

build/$(appname): $(sources)
	go build -ldflags "-X=main.build=$(build_version)" -v -o build/$(appname) cmd/$(appname)/main.go

test: go.test

clean:
	rm -rf build/*
	rm -rf postisto

install: build/$(appname)
	cp build/$(appname) $(GOBIN)/postisto

uninstall:
	rm -f $(GOBIN)/postisto

fmt:
	go fmt ./...

go.test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

go.get:
	go mod download

vendor-update: go.mod
	GO111MODULE=on go get -u=patch

vendor: go.mod
	GO111MODULE=on go mod vendor

docker-build:
	$(call build,linux,amd64,)
	docker build -t docker.pkg.github.com/arnisoph/postisto/linux:$(artifact_version) .
	make clean

docker-release: docker-build
	docker push docker.pkg.github.com/arnisoph/postisto/linux:$(artifact_version)

release: build docker-release


##### LINUX #####
linux: build/$(appname)-$(artifact_version).linux-amd64.tar.gz build/$(appname)-$(artifact_version).linux-arm7.tar.gz

build/$(appname)-$(artifact_version).linux-amd64.tar.gz: $(sources)
	$(call build,linux,amd64,)
	$(call tar,linux,amd64)

build/$(appname)-$(artifact_version).linux-arm7.tar.gz: $(sources)
	$(call build,linux,arm,,7)
	$(call tar,linux,arm7)


##### DARWIN (MAC) #####
darwin: build/$(appname)-$(artifact_version).darwin-amd64.tar.gz

build/$(appname)-$(artifact_version).darwin-amd64.tar.gz: $(sources)
	$(call build,darwin,amd64,)
	$(call tar,darwin,amd64)


##### WINDOWS #####
windows: build/$(appname)-$(artifact_version).windows-amd64.zip build/$(appname)-$(artifact_version).windows-arm7.zip

build/$(appname)-$(artifact_version).windows-amd64.zip: $(sources)
	$(call build,windows,amd64,.exe,)
	$(call zip,windows,amd64,.exe)

build/$(appname)-$(artifact_version).windows-arm7.zip: $(sources)
	$(call build,windows,arm,.exe,7)
	$(call zip,windows,arm7,.exe)
