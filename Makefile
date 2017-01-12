GOBUILD = GOOS="linux" go build -o
QUERY_DIR = ../../query
USER_DIR = ../../user
COMMON = src/common/common.go

.PHONY:
	build

all: query/install user/download user/share

query/install: src/install/main.go $(COMMON)
	cd src/install && $(GOBUILD) $(QUERY_DIR)/install

user/download: src/download/main.go $(COMMON)
	cd src/download && $(GOBUILD) $(USER_DIR)/download

user/share: src/share/main.go $(COMMON)
	cd src/share && $(GOBUILD) $(USER_DIR)/share
