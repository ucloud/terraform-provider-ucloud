SWEEP?=ch-bj2
TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=ucloud
WEBSITE_REPO=github.com/hashicorp/terraform-website

default: build

build: fmtcheck
	go install

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test $(TEST) -v -sweep=$(SWEEP) $(SWEEPARGS)

test: fmtcheck
	go test $(TEST) -timeout=30s -parallel=32

testacc: fmtcheck
	TF_ACC=1 go test -cover $(TEST) -v $(TESTARGS) -timeout 120m -parallel=32

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w -s $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

vendor-status:
	@dep status

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

.PHONY: build sweep test testacc vet fmt fmtcheck errcheck vendor-status test-compile website website-test

all: mac windows linux

dev: clean fmt
	@chmod +x scripts/devinit.sh
	@sh ./scripts/devinit.sh

clean:
	rm -rf bin/*

mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/terraform-provider-ucloud
	chmod +x bin/terraform-provider-ucloud
	cd bin/ && tar czvf ./terraform-provider-ucloud_darwin-amd64.tgz ./terraform-provider-ucloud
	rm -rf ./bin/terraform-provider-ucloud

windows:
	GOOS=windows GOARCH=amd64 go build -o bin/terraform-provider-ucloud.exe
	chmod +x bin/terraform-provider-ucloud.exe
	cd bin/ && tar czvf ./terraform-provider-ucloud_windows-amd64.tgz ./terraform-provider-ucloud.exe
	rm -rf ./bin/terraform-provider-ucloud.exe

linux:
	GOOS=linux GOARCH=amd64 go build -o bin/terraform-provider-ucloud
	chmod +x bin/terraform-provider-ucloud
	cd bin/ && tar czvf ./terraform-provider-ucloud_linux-amd64.tgz ./terraform-provider-ucloud
	rm -rf ./bin/terraform-provider-ucloud
