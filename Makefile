GO      ?= go
UPX_BIN ?= upx
BIN     := openwrt-monitor-tui
DIST    := dist
VERSION ?= dev
FLAGS   := -trimpath -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: all release host mipsle mips 386 amd64 pack fmt vet test clean

all: mipsle mips 386 amd64

## 发布链：编译 + UPX 压缩（CI 与 tag 发布用）
release: all pack

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

## UPX 压缩发布产物（约 9.2MB -> 3MB）。upx 不存在或单个文件压缩失败都只
## 提示（错误落在 dist/<name>.upx.log），不让 CI 变红；重复运行安全（已压
## 缩的自动跳过）。调试产物用 make all。
pack:
	@command -v $(UPX_BIN) >/dev/null 2>&1 || { echo "upx not found, skip packing"; exit 0; }
	@rm -f $(DIST)/$(BIN)-*.upx.log
	@for f in $(DIST)/$(BIN)-*; do \
		case "$$f" in *.upx|*.upx.log) continue;; esac; \
		if ! $(UPX_BIN) --best "$$f" > "$$f.upx.log" 2>&1; then \
			if grep -q AlreadyPacked "$$f.upx.log"; then \
				rm -f "$$f.upx.log"; \
			else \
				echo "skip $$(basename $$f) (see $$(basename $$f).upx.log)"; \
			fi; \
		else \
			rm -f "$$f.upx.log"; \
		fi; \
	done

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

clean:
	rm -rf $(DIST) $(BIN)
