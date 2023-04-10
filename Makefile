HOME = $(shell pwd)
BIN = $(shell basename ${HOME})

clean:
	rm -f $(BIN)

gen: 
	buf generate --path ./api -o ${HOME}/pkg

go-vendor:
	go mod tidy && go mod vendor

build: clean go-vendor
	env GOOS=linux GOARCH=arm go build -mod vendor -o $(BIN)

docker-build:
	docker build --no-cache --progress=plain -t signaler:1.0 -f Dockerfile .

docker-compose:
	docker-compose up --detach --build

run-local:
	go run main.go start --config config.dev.toml --env .env

rsync: build
	rsync -avz futures-signal *.toml runserver admin@54.179.74.34/home/admin/server/futures-signal

dockerhub:
	docker tag anvh2/futures-signal:v1.0.0 anvh2/futures-signal:v1.0.0
	docker push anvh2/futures-signal:v1.0.0