VERSION := $(shell git describe --tags --always --long)

build:
	GOOS=linux GOARCH=mipsle go build -o transip
	scp transip/router admin@192.168.1.1:~/

build-plugin:
	if [ ! -d "./build/plugins" ]; then \
  		mkdir -p ./build/plugins; \
  	fi
	find ./plugins \
		-maxdepth 1 \
		-mindepth 1 \
		-type d \
		-exec \
			go build -o build/plugins -ldflags '-w -s -X main.Version=$(VERSION)' -buildmode=plugin {} \
		\;