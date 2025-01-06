# Specify the GOOS and GOARCH variables
export CGO_ENABLED=0
export GOARCH=amd64

# Set to x84 arch for 32-bit builds
ifeq ($(MAKECMDGOALS), win32)
	export GOARCH=386
endif

# Remove debug information for release builds
ifeq ($(MAKECMDGOALS), debug)
	OUTPUT_DIR=bin/debug
	LDFLAGS=""
else 
	LDFLAGS="-s -w"
	OUTPUT_DIR=bin/release
endif

# Set output file and tags based on target game
ifeq ($(TARGET), ssl)
	ENTRYPOINT=./cmd/ssl
	OUTPUT_FILE=$(OUTPUT_DIR)/ssl_scraper
	TAGS=ssl
else ifeq ($(TARGET), ffo)
	ENTRYPOINT=./cmd/ffo
	OUTPUT_FILE=$(OUTPUT_DIR)/ffo_scraper
	TAGS=ffo
endif

# Append .exe to output file for Windows builds
ifeq ($(findstring win,$(MAKECMDGOALS)),win)
  OUTPUT_FILE := $(OUTPUT_FILE).exe
endif

BUILDCMD=go build -ldflags=$(LDFLAGS) -tags $(TAGS) -o $(OUTPUT_FILE) $(ENTRYPOINT)

debug:
	@echo "Building debug..."
	$(BUILDCMD)

win:
	@echo "Building for Windows (x64)..."
	GOOS=windows $(BUILDCMD)

win32:
	@echo "Building for Windows (x86)..."
	GOOS=windows $(BUILDCMD)

linux:
	@echo "Building for Linux..."
	GOOS=linux $(BUILDCMD)

mac:
	@echo "Building for Mac..."
	GOOS=darwin $(BUILDCMD)