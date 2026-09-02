SWEEP?=cn-bj2,cn-sh2
TEST?=./...
PRODUCT?=
PRODUCT_DIRS:=$(wildcard products/*/)
PRODUCTS:=$(sort $(notdir $(patsubst %/,%,$(PRODUCT_DIRS))))
PRODUCT_NAME:=$(if $(word 2,$(value PRODUCT)),,$(filter $(PRODUCTS),$(value PRODUCT)))
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=ucloud
WEBSITE_REPO=github.com/hashicorp/terraform-website

unexport PRODUCT

default: build

build: fmtcheck
	go install

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test $(TEST) -v -sweep=$(SWEEP) $(SWEEPARGS)

test: fmtcheck
	go test $(TEST) -timeout=30s -parallel=32

compat: fmtcheck
	go test ./internal/product ./internal/productcatalog ./internal/productownership ./internal/productownershipsync ./internal/providercompat -count=1
	go test ./ucloud ./products/... -run '^(TestProvider$$|TestProviderContract$$|TestProductClient|TestRegistration|TestBucket|TestUpgradeFixtureSyntax$$)' -count=1

testacc: fmtcheck
	@set -eu; \
		product="$(PRODUCT_NAME)"; \
		if [ -z "$$product" ]; then \
			echo "ERROR: Set PRODUCT to one of: $(PRODUCTS)" >&2; \
			exit 2; \
		fi; \
		if ! go test "./products/$$product" -list '^TestAcc' | grep -q '^TestAcc'; then \
			echo "ERROR: Product $$product has no acceptance tests" >&2; \
			exit 3; \
		fi; \
		TF_ACC=1 go test -cover "./products/$$product" -v -timeout 120m -parallel=32

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

.PHONY: build sweep test compat testacc vet fmt fmtcheck errcheck vendor-status test-compile website website-test

all: mac windows linux

dev: clean fmt
	@chmod +x scripts/devinit.sh
	@bash ./scripts/devinit.sh

clean:
	rm -rf bin/*

mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/terraform-provider-ucloud
	chmod +x bin/terraform-provider-ucloud
	cd bin/ && tar czvf ./terraform-provider-ucloud_darwin-amd64.tgz ./terraform-provider-ucloud
	rm -rf ./bin/terraform-provider-ucloud
mac-arm:
	GOOS=darwin GOARCH=arm64 go build -o bin/terraform-provider-ucloud
	chmod +x bin/terraform-provider-ucloud
	cd bin/ && tar czvf ./terraform-provider-ucloud_darwin-arm64.tgz ./terraform-provider-ucloud
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
