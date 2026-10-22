GO      ?= go
BIN     := openwrt-monitor-tui
DIST    := dist
VERSION ?= dev
FLAGS   := -trimpath -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: all host mipsle mips 386 amd64 fmt vet test clean

all: mipsle mips 386 amd64

## 小端 MIPS (ramips/mt7621、mt76x8：K2P、Newifi3、小米路由等)
mipsle:
	CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat $(GO) build $(FLAGS) -o $(DIST)/$(BIN)-mipsle .

## 大端 MIPS (ath79/ar71xx：多数老款 TP-Link、网件)
mips:
	CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat $(GO) build $(FLAGS) -o $(DIST)/$(BIN)-mips .

## OpenWrt 19.07 的 x86 默认是 32 位
386:
	CGO_ENABLED=0 GOOS=linux GOARCH=386 $(GO) build $(FLAGS) -o $(DIST)/$(BIN)-386 .

amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(FLAGS) -o $(DIST)/$(BIN)-amd64 .

host:
	$(GO) build -o $(BIN) .

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

clean:
	rm -rf $(DIST) $(BIN)
