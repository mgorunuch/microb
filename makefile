all: generate-build-commands build-all

generate-build-commands:
	@./scripts/create-build-commands.sh

GREEN=[32m
BLUE=[34m
RESET=[0m

# -- BUILD_COMMANDS START --
# Auto-generated build commands
build-all: build_alienvault_passivedns build_binary_edge build_certspotter build_commoncrawl build_google_custom_search build_web_archive 


build_alienvault_passivedns:
	@echo "$(BLUE)Building $(GREEN)alienvault_passivedns$(RESET)"
	@go build -o bin/alienvault_passivedns app/commands/alienvault_passivedns/main.go

build_binary_edge:
	@echo "$(BLUE)Building $(GREEN)binary_edge$(RESET)"
	@go build -o bin/binary_edge app/commands/binary_edge/main.go

build_certspotter:
	@echo "$(BLUE)Building $(GREEN)certspotter$(RESET)"
	@go build -o bin/certspotter app/commands/certspotter/main.go

build_commoncrawl:
	@echo "$(BLUE)Building $(GREEN)commoncrawl$(RESET)"
	@go build -o bin/commoncrawl app/commands/commoncrawl/main.go

build_google_custom_search:
	@echo "$(BLUE)Building $(GREEN)google_custom_search$(RESET)"
	@go build -o bin/google_custom_search app/commands/google_custom_search/main.go

build_web_archive:
	@echo "$(BLUE)Building $(GREEN)web_archive$(RESET)"
	@go build -o bin/web_archive app/commands/web_archive/main.go

# -- BUILD_COMMANDS END --

