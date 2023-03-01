SOURCES := $(shell find . -type f -name '*.go')
BIN_PATH := $(shell go env GOBIN)
BIN := $(BIN_PATH)/waktu-solat
PACKAGE_FILE := WaktuSolat.alfredworkflow
FILES := $(BIN) alfred-resources/info.plist alfred-resources/icon.png

build: $(BIN)

package: $(PACKAGE_FILE)

$(PACKAGE_FILE): $(FILES)
	zip -j "$@" $^

$(BIN): $(SOURCES)
	CGO_ENABLED=1 go build -ldflags="-s -w" -o $(BIN)
#	upx --best --lzma $(BIN)

clean:
	-rm $(BIN)

link:
