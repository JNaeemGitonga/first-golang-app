.PHONY: deps clean build

deps:
	dep ensure

clean: 
	rm -rf ./aahs-backend/aahs-backend
	
recompile:
	GOOS=linux GOARCH=amd64 go build -o aahs-backend/aahs-backend ./aahs-backend

supervise:
	supervisor --no-restart-on exit -e go -i bin --exec make -- recompile

start-local:
	sam local start-api

watch:
	make supervise & make start-local

build:
	deps
	make recompile