# Specify the GOOS and GOARCH variables
export CGO_ENABLED=0
export GOARCH=amd64
export GOOS=darwin

# Remove debug information for release builds
ifeq ($(MAKECMDGOALS), build)
	OUTPUT_DIR=bin/debug
	LDFLAGS=""
else 
	LDFLAGS="-s -w"
	OUTPUT_DIR=bin/release
endif

build:
	@BUILD_DIR=$(OUTPUT_DIR)/$(GOOS)/$(GOARCH) && \
	OUTPUT_FILE=$$BUILD_DIR/$(TARGET)_scraper$(FILE_EXT) && \
	rm -rf $$BUILD_DIR && \
	go build -ldflags=$(LDFLAGS) -tags $(TARGET) -o $$OUTPUT_FILE ./cmd/$(TARGET)

package:
	@PKG_DIR=$(OUTPUT_DIR)/$(GOOS)/$(GOARCH) && \
	echo "Packaging $$PKG_DIR..." && \
	if [ "$(GOOS)" = "linux" ]; then \
		tar -czvf "$(TARGET)_scraper.$(GOOS)_$(GOARCH).tgz" -C "$$PKG_DIR" .; \
	else \
		zip -rj "$$PKG_DIR/$(TARGET)_scraper.$(GOOS)_$(GOARCH).zip" "$$PKG_DIR"; \
	fi

win:
	@echo "Building for Windows (x64)..."
	@$(MAKE) FILE_EXT=.exe GOOS=windows build package

win32:
	@echo "Building for Windows (x86)..."
	@$(MAKE) FILE_EXT=.exe GOOS=windows GOARCH=386 build package

linux:
	@echo "Building for Linux..."
	@$(MAKE) GOOS=linux build package

mac:
	@echo "Building for MacOS..."
	@$(MAKE) GOOS=darwin build package