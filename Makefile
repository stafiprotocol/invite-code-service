PROJECTNAME=$(shell basename "$(PWD)")
.PHONY: help run build install

all:build

get:
	@echo "  >  \033[32mDownloading & Installing all the modules...\033[0m "
	go mod tidy && go mod download

fmt:
	@echo "  >  \033[32mfmt...\033[0m "
	go fmt ./...

build:
	@echo "  >  \033[32mBuilding binary...\033[0m "
	env GOARCH=amd64 go build -o ./build/invite-code-serviced


swagger:
	@echo "  >  \033[32mBuilding swagger docs...\033[0m "
	swag init --parseDependency
	

clean:
	rm -rf build/
