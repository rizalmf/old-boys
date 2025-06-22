## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# run/game: Running desktop
.PHONY: run/game
run/game:
	go run main.go

# run/web: Running web
.PHONY: run/web
run/web:
	C:/go/bin/wasmserve.exe -http=":8083" -allow-origin="*"