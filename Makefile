GO                ?= go
BINARY            ?= a2s
OUTPUT_DIR        ?= build
CGO_ENABLED       ?= 0
GOFLAGS           ?= -buildvcs=auto -trimpath
TEST_GOFLAGS      ?= -tags forceposix
LDFLAGS           ?= -s -w
GOWORK            ?= off
LANG              ?= C

LINTER            ?= golangci-lint
LINTER_VERSION    ?= v2.13.2
ALIGNER           ?= betteralign
ALIGNER_VERSION   ?= v0.15.1
VULNCHECK         ?= govulncheck
VULNCHECK_VERSION ?= v1.8.0
WINRES            ?= go-winres
WINRES_VERSION    ?= v0.3.3
WINRES_OUT        ?= ./cmd/$(BINARY)/rsrc
BENCHSTAT         ?= benchstat
RUMDL             ?= rumdl

RELEASE_MATRIX    ?= \
	linux/amd64 linux/arm64 \
	darwin/amd64 darwin/arm64 \
	windows/amd64 windows/arm64

MODULE_PATH       := $(shell $(GO) list -m -f '{{.Path}}')
NATIVE_GOOS       := $(shell $(GO) env GOOS)
NATIVE_GOARCH     := $(shell $(GO) env GOARCH)
BUILD_GOOS        ?= $(NATIVE_GOOS)
BUILD_GOARCH      ?= $(NATIVE_GOARCH)
BUILD_EXTENSION   := $(if $(filter $(BUILD_GOOS),windows),.exe,)
VERSION           := $(shell git describe --tags --abbrev=0 2>/dev/null || printf v0.0.0)
COMMIT            := $(shell git rev-parse HEAD 2>/dev/null || printf unknown)
DATE              := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
URL               := https://$(MODULE_PATH)
LDFLAGS_X         := \
	-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(COMMIT)' \
	-X 'main._buildTime=$(DATE)' \
	-X 'main.repositoryURL=$(URL)'

RACE ?= 0
ifeq ($(RACE),1)
	EXTRA_BUILD_FLAGS := -race
endif

BENCH_COUNT       ?= 6
BENCH_REF         ?= bench_baseline.txt
FUZZ_TIME         ?= 10s
FUZZ_TARGETS      ?= \
	./internal/a2srules:Parse \
	./internal/wire:DecoderCursorInvariants \
	./pkg/a2s:ParseChallenge \
	./pkg/a2s:ParseChallengeResponse \
	./pkg/a2s:DecodeInfoSource \
	./pkg/a2s:DecodeInfoGoldSource \
	./pkg/a2s:ParsePlayers \
	./pkg/a2s:ParseTheShipPlayers \
	./pkg/a2s:ParseRules \
	./pkg/a2s:ParsePacketHeaders \
	./pkg/a2s:DecompressBzip2 \
	./pkg/a2s:CreateHeader \
	./pkg/a2s:BinaryChallengeRoundTrip \
	./pkg/a2s/server:DecodeRequest \
	./pkg/a2s/server:DecodePacket \
	./pkg/a2s/server:EncodeInfo \
	./pkg/a2s/server:EncodePlayers \
	./pkg/a2s/server:EncodeRules \
	./pkg/a2s/server:Packetizers \
	./pkg/a2s/server:ChallengeGate \
	./pkg/a3sb:ReadA3SB \
	./pkg/a3sb:BuildPageEnvelope \
	./pkg/a3sb:ParseAutomaticRules \
	./pkg/a3sb:ParseRulesDayZ \
	./pkg/keywords:ParseKeywords \
	./pkg/keywords:ParseCoordinates

export GOWORK
export LANG

.PHONY: clean build compile release

clean:
	rm -rf $(OUTPUT_DIR)

build: clean $(if $(filter windows,$(BUILD_GOOS)),winres)
	@mkdir -p $(OUTPUT_DIR)
	@echo ">> build: $(BINARY)$(BUILD_EXTENSION) ($(BUILD_GOOS)/$(BUILD_GOARCH))"
	GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) \
	GOWORK=$(GOWORK) CGO_ENABLED=$(CGO_ENABLED) \
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS) $(LDFLAGS_X)" $(EXTRA_BUILD_FLAGS) \
	-o $(OUTPUT_DIR)/$(BINARY)$(BUILD_EXTENSION) ./cmd/$(BINARY)

release: clean winres
	@mkdir -p $(OUTPUT_DIR)
	@set -e; \
	for target in $(RELEASE_MATRIX); do \
		goos=$${target%%/*}; \
		goarch=$${target##*/}; \
		ext=$$( [ $$goos = "windows" ] && echo ".exe" || echo "" ); \
		out="$(OUTPUT_DIR)/$(BINARY)-$${goos}-$${goarch}$$ext"; \
		echo ">> build $$out"; \
		GOOS=$$goos GOARCH=$$goarch GOWORK=$(GOWORK) CGO_ENABLED=$(CGO_ENABLED) \
			$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS) $(LDFLAGS_X)" -o $$out ./cmd/$(BINARY); \
	done

.PHONY: check ci

check: verify tidy fmt vet vulncheck lint-fix align-fix test-short test-race-short generate-docs markdown-lint
ci: download generate-check tools-ci verify tidy-check fmt-check vet vulncheck lint align test-short

.PHONY: test test-short test-race test-race-short fuzz

test:
	$(GO) test $(TEST_GOFLAGS) ./...

test-short:
	$(GO) test $(TEST_GOFLAGS) -short ./...

test-race:
	CGO_ENABLED=1 $(GO) test $(TEST_GOFLAGS) -race ./...

test-race-short:
	CGO_ENABLED=1 $(GO) test $(TEST_GOFLAGS) -short -race ./...

fuzz:
	@set -e; \
	for target in $(FUZZ_TARGETS); do \
		echo "fuzz target: $${target##*:} from $${target%%:*}"; \
		$(GO) test $(TEST_GOFLAGS) $${target%%:*} -run='^$$' -fuzz='^Fuzz'$${target##*:}'$$' -fuzztime=$(FUZZ_TIME); \
	done

.PHONY: bench bench-fast bench-reset

bench:
	@tmp=$$(mktemp); \
	$(GO) test $(TEST_GOFLAGS) ./... -run=^$$ -bench 'Benchmark' -benchmem -count=$(BENCH_COUNT) | tee "$$tmp"; \
	if [ -f "$(BENCH_REF)" ]; then \
		$(BENCHSTAT) "$(BENCH_REF)" "$$tmp"; \
	else \
		cp "$$tmp" "$(BENCH_REF)" && echo "Baseline saved to $(BENCH_REF)"; \
	fi; \
	rm -f "$$tmp"

bench-fast:
	$(GO) test $(TEST_GOFLAGS) ./... -run=^$$ -bench 'Benchmark' -benchmem

bench-reset:
	rm -f "$(BENCH_REF)"

.PHONY: download verify vet tidy tidy-check fmt fmt-check vulncheck lint lint-fix align align-fix

download:
	$(GO) mod download

verify:
	$(GO) mod verify

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

tidy-check:
	@$(GO) mod tidy
	@git diff --stat --exit-code -- go.mod go.sum || ( \
		echo "go mod tidy: repository is not tidy"; \
		exit 1; \
	)

fmt:
	$(GO) fmt ./...

fmt-check:
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		echo "$$files"; \
		echo "gofmt: files need formatting"; \
		exit 1; \
	fi

vulncheck:
	$(VULNCHECK) ./...

lint:
	$(LINTER) run ./...

lint-fix:
	$(LINTER) run --fix ./...

align:
	$(ALIGNER) ./...

align-fix:
	-$(ALIGNER) -apply ./...
	$(ALIGNER) ./...

.PHONY: generate-docs generate-check

generate-docs:
	LANG=en $(GO) run ./cmd/$(BINARY) docs md CLI.md --program-name $(BINARY) \
		--style posix --template table

generate-check: generate-docs
	@git diff --stat --exit-code -- CLI.md || ( \
		echo "CLI docs are out of date; run 'make generate-docs' and commit changes"; \
		exit 1; \
	)

.PHONY: winres

winres:
	$(WINRES) make \
	--in winres/manifest.json \
	--arch amd64,arm64 \
	--out $(WINRES_OUT) \
	--product-version "$(VERSION)" \
	--file-version "$(VERSION)"

.PHONY: tools tools-ci tools-build \
	tool-golangci-lint tool-betteralign tool-govulncheck tool-winres tool-benchstat

tools: tool-golangci-lint tool-betteralign tool-govulncheck tool-winres tool-benchstat
tools-ci: tool-golangci-lint tool-betteralign tool-govulncheck

tool-golangci-lint:
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(LINTER_VERSION)

tool-betteralign:
	$(GO) install github.com/dkorunic/betteralign/cmd/betteralign@$(ALIGNER_VERSION)

tool-govulncheck:
	$(GO) install golang.org/x/vuln/cmd/govulncheck@$(VULNCHECK_VERSION)

tool-winres:
	$(GO) install github.com/tc-hib/go-winres@$(WINRES_VERSION)

tool-benchstat:
	$(GO) install golang.org/x/perf/cmd/benchstat@latest

.PHONY: release-notes

release-notes:
	@awk '\
	/^<!--/,/^-->/ { next } \
	/^## \[[0-9]+\.[0-9]+\.[0-9]+\]/ { if (found) exit; found=1; next } \
	found { \
		if (/^## \[/) { exit } \
		if (/^$$/) { flush(); print; next } \
		if (/^\* / || /^- /) { flush(); buf=$$0; next } \
		if (/^###/ || /^\[/) { flush(); print; next } \
		sub(/^[ \t]+/, ""); sub(/[ \t]+$$/, ""); \
		if (buf != "") { buf = buf " " $$0 } else { buf = $$0 } \
		next \
	} \
	function flush() { if (buf != "") { print buf; buf = "" } } \
	END { flush() } \
	' CHANGELOG.md

.PHONY: markdown-lint markdown-fix

define run-rumdl
	@if command -v $(RUMDL) &>/dev/null; then \
		$(RUMDL) $(1); \
	else \
		echo "WARN: $(RUMDL) not found; skipping markdown lint."; \
		echo 'WARN: Install it https://github.com/rvben/rumdl'; \
	fi
endef

markdown-lint:
	$(call run-rumdl,check)

markdown-fix:
	$(call run-rumdl,check --fix)
