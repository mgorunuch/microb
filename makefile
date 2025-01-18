all: generate-build-commands build-all

generate-build-commands:
	@./scripts/create-build-commands.sh

GREEN=[32m
BLUE=[34m
RESET=[0m

# -- BUILD_COMMANDS START --
# Auto-generated build commands
build-all: build_alienvault_passivedns build_binary_edge build_certspotter build_chrome_visit_html build_commoncrawl build_crt_sh build_extract_domains build_google_custom_search build_itterate_yasss build_itterate_yasss_status build_link_extractor build_migrate build_open_chrome build_store_domains build_store_links build_unique_lines build_web_archive 


build_alienvault_passivedns:
	@echo "$(BLUE)Building $(GREEN)alienvault_passivedns$(RESET)"
	@go build -o bin/alienvault_passivedns app/commands/alienvault_passivedns/main.go

build_binary_edge:
	@echo "$(BLUE)Building $(GREEN)binary_edge$(RESET)"
	@go build -o bin/binary_edge app/commands/binary_edge/main.go

build_certspotter:
	@echo "$(BLUE)Building $(GREEN)certspotter$(RESET)"
	@go build -o bin/certspotter app/commands/certspotter/main.go

build_chrome_visit_html:
	@echo "$(BLUE)Building $(GREEN)chrome_visit_html$(RESET)"
	@go build -o bin/chrome_visit_html app/commands/chrome_visit_html/main.go

build_commoncrawl:
	@echo "$(BLUE)Building $(GREEN)commoncrawl$(RESET)"
	@go build -o bin/commoncrawl app/commands/commoncrawl/main.go

build_crt_sh:
	@echo "$(BLUE)Building $(GREEN)crt_sh$(RESET)"
	@go build -o bin/crt_sh app/commands/crt_sh/main.go

build_extract_domains:
	@echo "$(BLUE)Building $(GREEN)extract_domains$(RESET)"
	@go build -o bin/extract_domains app/commands/extract_domains/main.go

build_google_custom_search:
	@echo "$(BLUE)Building $(GREEN)google_custom_search$(RESET)"
	@go build -o bin/google_custom_search app/commands/google_custom_search/main.go

build_itterate_yasss:
	@echo "$(BLUE)Building $(GREEN)itterate_yasss$(RESET)"
	@go build -o bin/itterate_yasss app/commands/itterate_yasss/main.go

build_itterate_yasss_status:
	@echo "$(BLUE)Building $(GREEN)itterate_yasss_status$(RESET)"
	@go build -o bin/itterate_yasss_status app/commands/itterate_yasss_status/main.go

build_link_extractor:
	@echo "$(BLUE)Building $(GREEN)link_extractor$(RESET)"
	@go build -o bin/link_extractor app/commands/link_extractor/main.go

build_migrate:
	@echo "$(BLUE)Building $(GREEN)migrate$(RESET)"
	@go build -o bin/migrate app/commands/migrate/main.go

build_open_chrome:
	@echo "$(BLUE)Building $(GREEN)open_chrome$(RESET)"
	@go build -o bin/open_chrome app/commands/open_chrome/main.go

build_store_domains:
	@echo "$(BLUE)Building $(GREEN)store_domains$(RESET)"
	@go build -o bin/store_domains app/commands/store_domains/main.go

build_store_links:
	@echo "$(BLUE)Building $(GREEN)store_links$(RESET)"
	@go build -o bin/store_links app/commands/store_links/main.go

build_unique_lines:
	@echo "$(BLUE)Building $(GREEN)unique_lines$(RESET)"
	@go build -o bin/unique_lines app/commands/unique_lines/main.go

build_web_archive:
	@echo "$(BLUE)Building $(GREEN)web_archive$(RESET)"
	@go build -o bin/web_archive app/commands/web_archive/main.go

# -- BUILD_COMMANDS END --

run: run_alien_vault_passive_dns run_binary_edge run_certspotter run_commoncrawl

run_alien_vault_passive_dns:
	@cat mock/domains.txt | ./bin/alienvault_passivedns

run_binary_edge:
	@cat mock/domains.txt | ./bin/binary_edge

run_certspotter:
	@cat mock/domains.txt | ./bin/certspotter

run_commoncrawl:
	@cat mock/domains.txt | ./bin/commoncrawl
