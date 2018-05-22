VERSION = 0.0.1

APP      := image-palette-tools
PACKAGES := $(shell go list -f {{.Dir}} ./...)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))

.PHONY: clean release release-ci release-manual README.md

clean: 
	rm -rf binaries/
	rm -rf release/

release-ci: README.md zip

release: README.md
	git reset
	git add README.md
	git add Makefile
	git commit -am "Release $(VERSION)" || true
	git tag "$(VERSION)"
	git push
	git push origin "$(VERSION)"

# go get -u github.com/github/hub
release-manual: README.md zip
	git push
	hub release create $(VERSION) -m "$(VERSION)" -a release/extract-palette_$(VERSION)_osx_x86_64.tar.gz -a release/extract-palette_$(VERSION)_windows_x86_64.zip -a release/extract-palette_$(VERSION)_linux_x86_64.tar.gz -a release/extract-palette_$(VERSION)_osx_x86_32.tar.gz -a release/extract-palette_$(VERSION)_windows_x86_32.zip -a release/extract-palette_$(VERSION)_linux_x86_32.tar.gz -a release/extract-palette_$(VERSION)_linux_arm64.tar.gz -a release/cluster-by-palette_$(VERSION)_osx_x86_64.tar.gz -a release/cluster-by-palette_$(VERSION)_windows_x86_64.zip -a release/cluster-by-palette_$(VERSION)_linux_x86_64.tar.gz -a release/cluster-by-palette_$(VERSION)_osx_x86_32.tar.gz -a release/cluster-by-palette_$(VERSION)_windows_x86_32.zip -a release/cluster-by-palette_$(VERSION)_linux_x86_32.tar.gz -a release/cluster-by-palette_$(VERSION)_linux_arm64.tar.gz

README.md:
	sed "s/\$${VERSION}/$(VERSION)/g;s/\$${APP}/$(APP)/g;" README.template.md > README.md

zip: release/extract-palette_$(VERSION)_osx_x86_64.tar.gz release/extract-palette_$(VERSION)_windows_x86_64.zip release/extract-palette_$(VERSION)_linux_x86_64.tar.gz release/extract-palette_$(VERSION)_osx_x86_32.tar.gz release/extract-palette_$(VERSION)_windows_x86_32.zip release/extract-palette_$(VERSION)_linux_x86_32.tar.gz release/extract-palette_$(VERSION)_linux_arm64.tar.gz release/cluster-by-palette_$(VERSION)_osx_x86_64.tar.gz release/cluster-by-palette_$(VERSION)_windows_x86_64.zip release/cluster-by-palette_$(VERSION)_linux_x86_64.tar.gz release/cluster-by-palette_$(VERSION)_osx_x86_32.tar.gz release/cluster-by-palette_$(VERSION)_windows_x86_32.zip release/cluster-by-palette_$(VERSION)_linux_x86_32.tar.gz release/cluster-by-palette_$(VERSION)_linux_arm64.tar.gz

binaries: binaries/osx_x86_64/extract-palette binaries/windows_x86_64/extract-palette.exe binaries/linux_x86_64/extract-palette binaries/osx_x86_32/extract-palette binaries/windows_x86_32/extract-palette.exe binaries/linux_x86_32/extract-palette binaries/osx_x86_64/cluster-by-palette binaries/windows_x86_64/cluster-by-palette.exe binaries/linux_x86_64/cluster-by-palette binaries/osx_x86_32/cluster-by-palette binaries/windows_x86_32/cluster-by-palette.exe binaries/linux_x86_32/cluster-by-palette

release/extract-palette_$(VERSION)_osx_x86_64.tar.gz: binaries/osx_x86_64/extract-palette
	mkdir -p release
	tar cfz release/extract-palette_$(VERSION)_osx_x86_64.tar.gz -C binaries/osx_x86_64 extract-palette
	
binaries/osx_x86_64/extract-palette: $(GOFILES)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/osx_x86_64/extract-palette ./cmd/extract-palette

release/extract-palette_$(VERSION)_windows_x86_64.zip: binaries/windows_x86_64/extract-palette.exe
	mkdir -p release
	cd ./binaries/windows_x86_64 && zip -r -D ../../release/extract-palette_$(VERSION)_windows_x86_64.zip extract-palette.exe
	
binaries/windows_x86_64/extract-palette.exe: $(GOFILES)
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/windows_x86_64/extract-palette.exe ./cmd/extract-palette

release/extract-palette_$(VERSION)_linux_x86_64.tar.gz: binaries/linux_x86_64/extract-palette
	mkdir -p release
	tar cfz release/extract-palette_$(VERSION)_linux_x86_64.tar.gz -C binaries/linux_x86_64 extract-palette
	
binaries/linux_x86_64/extract-palette: $(GOFILES)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/linux_x86_64/extract-palette ./cmd/extract-palette

release/extract-palette_$(VERSION)_osx_x86_32.tar.gz: binaries/osx_x86_32/extract-palette
	mkdir -p release
	tar cfz release/extract-palette_$(VERSION)_osx_x86_32.tar.gz -C binaries/osx_x86_32 extract-palette
	
binaries/osx_x86_32/extract-palette: $(GOFILES)
	GOOS=darwin GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -o binaries/osx_x86_32/extract-palette ./cmd/extract-palette

release/extract-palette_$(VERSION)_windows_x86_32.zip: binaries/windows_x86_32/extract-palette.exe
	mkdir -p release
	cd ./binaries/windows_x86_32 && zip -r -D ../../release/extract-palette_$(VERSION)_windows_x86_32.zip extract-palette.exe
	
binaries/windows_x86_32/extract-palette.exe: $(GOFILES)
	GOOS=windows GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -o binaries/windows_x86_32/extract-palette.exe ./cmd/extract-palette

release/extract-palette_$(VERSION)_linux_x86_32.tar.gz: binaries/linux_x86_32/extract-palette
	mkdir -p release
	tar cfz release/extract-palette_$(VERSION)_linux_x86_32.tar.gz -C binaries/linux_x86_32 extract-palette
	
binaries/linux_x86_32/extract-palette: $(GOFILES)
	GOOS=linux GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -o binaries/linux_x86_32/extract-palette ./cmd/extract-palette

release/extract-palette_$(VERSION)_linux_arm64.tar.gz: binaries/linux_arm64/extract-palette
	mkdir -p release
	tar cfz release/extract-palette_$(VERSION)_linux_arm64.tar.gz -C binaries/linux_arm64 extract-palette
	
binaries/linux_arm64/extract-palette: $(GOFILES)
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/linux_arm64/extract-palette ./cmd/extract-palette

release/cluster-by-palette_$(VERSION)_osx_x86_64.tar.gz: binaries/osx_x86_64/cluster-by-palette
	mkdir -p release
	tar cfz release/cluster-by-palette_$(VERSION)_osx_x86_64.tar.gz -C binaries/osx_x86_64 cluster-by-palette
	
binaries/osx_x86_64/cluster-by-palette: $(GOFILES)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/osx_x86_64/cluster-by-palette ./cmd/cluster-by-palette

release/cluster-by-palette_$(VERSION)_windows_x86_64.zip: binaries/windows_x86_64/cluster-by-palette.exe
	mkdir -p release
	cd ./binaries/windows_x86_64 && zip -r -D ../../release/cluster-by-palette_$(VERSION)_windows_x86_64.zip cluster-by-palette.exe
	
binaries/windows_x86_64/cluster-by-palette.exe: $(GOFILES)
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/windows_x86_64/cluster-by-palette.exe ./cmd/cluster-by-palette

release/cluster-by-palette_$(VERSION)_linux_x86_64.tar.gz: binaries/linux_x86_64/cluster-by-palette
	mkdir -p release
	tar cfz release/cluster-by-palette_$(VERSION)_linux_x86_64.tar.gz -C binaries/linux_x86_64 cluster-by-palette
	
binaries/linux_x86_64/cluster-by-palette: $(GOFILES)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/linux_x86_64/cluster-by-palette ./cmd/cluster-by-palette

release/cluster-by-palette_$(VERSION)_osx_x86_32.tar.gz: binaries/osx_x86_32/cluster-by-palette
	mkdir -p release
	tar cfz release/cluster-by-palette_$(VERSION)_osx_x86_32.tar.gz -C binaries/osx_x86_32 cluster-by-palette
	
binaries/osx_x86_32/cluster-by-palette: $(GOFILES)
	GOOS=darwin GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -o binaries/osx_x86_32/cluster-by-palette ./cmd/cluster-by-palette

release/cluster-by-palette_$(VERSION)_windows_x86_32.zip: binaries/windows_x86_32/cluster-by-palette.exe
	mkdir -p release
	cd ./binaries/windows_x86_32 && zip -r -D ../../release/cluster-by-palette_$(VERSION)_windows_x86_32.zip cluster-by-palette.exe
	
binaries/windows_x86_32/cluster-by-palette.exe: $(GOFILES)
	GOOS=windows GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -o binaries/windows_x86_32/cluster-by-palette.exe ./cmd/cluster-by-palette

release/cluster-by-palette_$(VERSION)_linux_x86_32.tar.gz: binaries/linux_x86_32/cluster-by-palette
	mkdir -p release
	tar cfz release/cluster-by-palette_$(VERSION)_linux_x86_32.tar.gz -C binaries/linux_x86_32 cluster-by-palette
	
binaries/linux_x86_32/cluster-by-palette: $(GOFILES)
	GOOS=linux GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -o binaries/linux_x86_32/cluster-by-palette ./cmd/cluster-by-palette

release/cluster-by-palette_$(VERSION)_linux_arm64.tar.gz: binaries/linux_arm64/cluster-by-palette
	mkdir -p release
	tar cfz release/cluster-by-palette_$(VERSION)_linux_arm64.tar.gz -C binaries/linux_arm64 cluster-by-palette
	
binaries/linux_arm64/cluster-by-palette: $(GOFILES)
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/linux_arm64/cluster-by-palette ./cmd/cluster-by-palette
