appname := postisto

sources = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
timestamp = $(shell date +"%Y.%m.%d")
gitrev = $(shell git rev-parse --short HEAD)
artifact_version = v$(timestamp)-$(gitrev)

build = GOOS=$(1) GOARCH=$(2) go build -trimpath -ldflags "-X=main.build=$(artifact_version)" -o build/$(appname)$(3) cmd/$(appname)/main.go
tar = cd build && tar -cvzf $(appname)-$(artifact_version).$(1)-$(2).tar.gz $(appname)$(3) && rm $(appname)$(3)
zip = cd build && zip $(appname)-$(artifact_version).$(1)-$(2).zip $(appname)$(3) && rm $(appname)$(3)

all: build test install

.PHONY: build test go.test clean fmt vendor-update vendor docker-build release github-release docker-release install version git-release

build: clean docker-build windows darwin linux

build/$(appname): $(sources)
	go build -ldflags "-X=main.build=$(artifact_version)" -v -o build/$(appname) cmd/$(appname)/main.go

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
vendor:
	go mod vendor
	cd cmd/postisto/ && go get -u=patch
	cd test/integration/ && go get -u=patch
	go mod tidy

docker-build: linux
	docker build --platform linux/amd64,linux/arm64 -t ghcr.io/arnisoph/postisto/linux:$(artifact_version) .
	make clean

docker-release: docker-build
	docker push ghcr.io/arnisoph/postisto/linux:$(artifact_version)

git-release:
	git tag $(artifact_version)

release: git-release build docker-release

version:
	@echo $(artifact_version)

##### LINUX #####
linux: build/$(appname)-$(artifact_version).linux-amd64.tar.gz build/$(appname)-$(artifact_version).linux-arm64.tar.gz

build/$(appname)-$(artifact_version).linux-amd64.tar.gz: $(sources)
	$(call build,linux,amd64,)
	$(call tar,linux,amd64)

build/$(appname)-$(artifact_version).linux-arm64.tar.gz: $(sources)
	$(call build,linux,arm64,)
	$(call tar,linux,arm64)


##### DARWIN (MAC) #####
darwin: build/$(appname)-$(artifact_version).darwin-amd64.tar.gz build/$(appname)-$(artifact_version).darwin-arm64.tar.gz

build/$(appname)-$(artifact_version).darwin-amd64.tar.gz: $(sources)
	$(call build,darwin,amd64,)
	$(call tar,darwin,amd64)

build/$(appname)-$(artifact_version).darwin-arm64.tar.gz: $(sources)
	$(call build,darwin,arm64,)
	$(call tar,darwin,arm64)


##### WINDOWS #####
windows: build/$(appname)-$(artifact_version).windows-amd64.zip build/$(appname)-$(artifact_version).windows-arm64.zip

build/$(appname)-$(artifact_version).windows-amd64.zip: $(sources)
	$(call build,windows,amd64,.exe,)
	$(call zip,windows,amd64,.exe)

build/$(appname)-$(artifact_version).windows-arm64.zip: $(sources)
	$(call build,windows,arm64,.exe,)
	$(call zip,windows,arm64,.exe)
