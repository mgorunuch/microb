build: build_alienvault_passivedns build_binary_edge build_google_custom_search build_certspotter

build_alienvault_passivedns:
	@echo "\x1b[34mBuilding \x1b[32malienvault_passivedns\x1b[0m"
	@go build -o bin/alienvault_passivedns app/commands/alienvault_passivedns/main.go

build_binary_edge:
	@echo "\x1b[34mBuilding \x1b[32mbinary_edge\x1b[0m"
	@go build -o bin/binary_edge app/commands/binary_edge/main.go

build_google_custom_search:
	@echo "\x1b[34mBuilding \x1b[32mgoogle_custom_search\x1b[0m"
	@go build -o bin/google_custom_search app/commands/google_custom_search/main.go

build_certspotter:
	@echo "\x1b[34mBuilding \x1b[32mcertspotter\x1b[0m"
	@go build -o bin/certspotter app/commands/certspotter/main.go
