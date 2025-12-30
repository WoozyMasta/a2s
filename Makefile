# ------------------------------- Settings ----------------------------------
RELEASE_MATRIX := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

CGO_ENABLED ?= 0
GOFLAGS     ?= -buildvcs=auto -trimpath
LDFLAGS     ?= -s -w
GOWORK      ?= off
GOFTAGS     ?= forceposix

NATIVE_GOOS      := $(shell go env GOOS)
NATIVE_GOARCH    := $(shell go env GOARCH)
NATIVE_EXTENSION := $(if $(filter $(NATIVE_GOOS),windows),.exe,)

BINARY     ?= a2s
PKG        ?= ./cmd/a2s
OUTPUT_DIR ?= build

# Tools
GO        ?= go
LINTER    ?= golangci-lint
ALIGNER   ?= betteralign
WINRES    ?= go-winres
CYCLO     ?= cyclonedx-gomod

# Optional race flag for native build: make build RACE=1
RACE ?= 0
ifeq ($(RACE),1)
	EXTRA_BUILD_FLAGS := -race
endif

# ----------------------------- Build metadata ------------------------------
MODULE  := $(shell go list -m)
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
VERSION_NO_V := $(patsubst v%,%,$(VERSION))
COMMIT  := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
URL     := https://$(MODULE)

LDFLAGS_X := \
	-X '$(MODULE)/internal/vars.Version=$(VERSION)' \
	-X '$(MODULE)/internal/vars.Commit=$(COMMIT)' \
	-X '$(MODULE)/internal/vars._buildTime=$(DATE)' \
	-X '$(MODULE)/internal/vars.URL=$(URL)'

# ---------------------------------------------------------------------------
.PHONY: all build release clean tidy download fmt verify vet tools lint align align-fix \
        sbom sbom-app sbom-bin release-notes _winres_patch _build_one _sbom_bin_one

all: tools check release

check: download tidy verify vet fmt lint align

clean:
	rm -rf $(OUTPUT_DIR)

# ------------------------------- Build -------------------------------------
build: clean
	@mkdir -p $(OUTPUT_DIR)
	@echo ">> building native: $(BINARY)$(NATIVE_EXTENSION)"
	GOOS=$(NATIVE_GOOS) GOARCH=$(NATIVE_GOARCH) \
	GOWORK=$(GOWORK) CGO_ENABLED=$(CGO_ENABLED) \
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS) $(LDFLAGS_X)" -tags "$(GOFTAGS)" $(EXTRA_BUILD_FLAGS) \
	-o $(OUTPUT_DIR)/$(BINARY)$(NATIVE_EXTENSION) $(PKG)
	@$(MAKE) _winres_patch GOOS=$(NATIVE_GOOS) GOARCH=$(NATIVE_GOARCH) BIN=$(BINARY) OUTEXT="$(NATIVE_EXTENSION)"
	@$(MAKE) _sbom_bin_one GOOS=$(NATIVE_GOOS) GOARCH=$(NATIVE_GOARCH) BIN=$(BINARY) OUTEXT="$(NATIVE_EXTENSION)"

release: clean
	@mkdir -p $(OUTPUT_DIR)
	@for target in $(RELEASE_MATRIX); do \
		goos=$${target%%/*}; \
		goarch=$${target##*/}; \
		ext=$$( [ $$goos = "windows" ] && echo ".exe" || echo "" ); \
		out="$(OUTPUT_DIR)/$(BINARY)-$${goos}-$${goarch}$$ext"; \
		echo ">> building $$out"; \
		GOOS=$$goos GOARCH=$$goarch \
		GOWORK=$(GOWORK) CGO_ENABLED=$(CGO_ENABLED) \
		$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS) $(LDFLAGS_X)" -tags "$(GOFTAGS)" \
		-o $$out $(PKG); \
		$(MAKE) --no-print-directory _winres_patch GOOS=$$goos GOARCH=$$goarch BIN=$(BINARY)-$${goos}-$${goarch} OUTEXT="$$ext"; \
		$(MAKE) --no-print-directory _sbom_bin_one GOOS=$$goos GOARCH=$$goarch BIN=$(BINARY)-$${goos}-$${goarch} OUTEXT="$$ext"; \
	done
	@$(MAKE) sbom-app

# ------------------------------- SBOM ---------------------------------------
sbom: sbom-app sbom-bin

sbom-app:
	@echo ">> SBOM (app)"
	$(CYCLO) app -json -packages -files -licenses \
		-output "$(OUTPUT_DIR)/$(BINARY).sbom.json" -main "$(PKG)"

sbom-bin:
	@echo ">> SBOM (bin native if exists)"
	@[ -f "$(OUTPUT_DIR)/$(BINARY)$(NATIVE_EXTENSION)" ] && \
		$(CYCLO) bin -json -output "$(OUTPUT_DIR)/$(BINARY)$(NATIVE_EXTENSION).sbom.json" \
			"$(OUTPUT_DIR)/$(BINARY)$(NATIVE_EXTENSION)" || true

_sbom_bin_one:
	@bin="$(OUTPUT_DIR)/$(BIN)$(OUTEXT)"; \
	if [ -f "$$bin" ]; then \
		echo ">> SBOM (bin) $$bin"; \
		$(CYCLO) bin -json -output "$$bin.sbom.json" "$$bin"; \
	fi

# ------------------------------ Win resources -------------------------------
_winres_patch:
	@if [ "$(GOOS)" = "windows" ]; then \
		winres_file="winres/winres.build.json"; \
		if [ ! -f "$$winres_file" ] && [ -f "winres/winres.json" ]; then \
			echo ">> updating winres version to $(VERSION_NO_V)"; \
			./winres/update.sh "$(VERSION_NO_V)"; \
		fi; \
		if [ ! -f "$$winres_file" ]; then \
			winres_file="winres/winres.json"; \
		fi; \
		if [ -f "$$winres_file" ]; then \
			echo ">> patch winres for $(BIN)$(OUTEXT)"; \
			$(WINRES) patch --no-backup --in "$$winres_file" "$(OUTPUT_DIR)/$(BIN)$(OUTEXT)"; \
		fi; \
	fi

# ------------------------------ Housekeeping --------------------------------
tidy:
	$(GO) mod tidy

download:
	$(GO) mod download

fmt:
	$(GO) fmt ./...

verify:
	$(GO) mod verify

vet:
	$(GO) vet ./...

tools:
	@echo ">> installing golangci-lint"
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@echo ">> installing betteralign"
	$(GO) install github.com/dkorunic/betteralign/cmd/betteralign@latest
	@echo ">> installing go-winres"
	$(GO) install github.com/tc-hib/go-winres@latest
	@echo ">> installing cyclonedx-gomod"
	$(GO) install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest

lint:
	$(LINTER) run ./...

align:
	$(ALIGNER) ./...

align-fix:
	$(ALIGNER) -apply ./...

release-notes:
	@awk '\
	/^<!--/,/^-->/ { next } \
	/^## \[[0-9]+\.[0-9]+\.[0-9]+\]/ { if (found) exit; found=1; next } found { print } \
	' CHANGELOG.md
