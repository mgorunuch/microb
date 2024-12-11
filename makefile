build: build_alienvault_passivedns build_binary_edge build_google_custom_search build_certspotter

GREEN=\033[32m
BLUE=\033[34m
RESET=\033[0m

build_alienvault_passivedns:
	@echo "$(BLUE)Building $(GREEN)alienvault_passivedns$(RESET)"
	@go build -o bin/alienvault_passivedns app/commands/alienvault_passivedns/main.go

build_binary_edge:
	@echo "$(BLUE)Building $(GREEN)binary_edge$(RESET)"
	@go build -o bin/binary_edge app/commands/binary_edge/main.go

build_google_custom_search:
	@echo "$(BLUE)Building $(GREEN)google_custom_search$(RESET)"
	@go build -o bin/google_custom_search app/commands/google_custom_search/main.go

build_certspotter:
	@echo "$(BLUE)Building $(GREEN)certspotter$(RESET)"
	@go build -o bin/certspotter app/commands/certspotter/main.go
