HOME = $(shell pwd)
BIN = $(shell basename ${HOME})

clean:
	rm -f $(BIN)

gen: 
	buf generate --path ./api -o ${HOME}/pkg

go-vendor:
	go mod tidy && go mod vendor

build: clean go-vendor
	GOOS=linux GOARCH=amd64 go build -o $(BIN)

docker-build:
	docker build --no-cache --progress=plain -t signaler:1.0 -f Dockerfile .

deploy:
	go run main.go start --config config.dev.toml --env .env