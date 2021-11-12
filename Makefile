GO 	    ?= GO111MODULE=on go
VERSION ?= latest
GITHASH ?= -

LDFLAGS := -s -w
LDFLAGS += -X "github.com/go-sdk/lib/app.VERSION=$(VERSION)"
LDFLAGS += -X "github.com/go-sdk/lib/app.GITHASH=$(GITHASH)"
LDFLAGS += -X "github.com/go-sdk/lib/app.BUILT=$(shell date +%FT%T%z)"

build:
	@$(MAKE) tidy
	CGO_ENABLED=0 $(GO) build -ldflags '$(LDFLAGS)' -o bin/app .

run:
	@$(MAKE) tidy
	CGO_ENABLED=1 $(GO) run -race -ldflags '$(LDFLAGS)' .

tidy:
	$(GO) mod tidy

upx:
	upx bin/app
