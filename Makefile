run:
	go run main.go

build:
	go build -o lighthouse-api

full-build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o lighthouse-api

doc:
	swag init -g ./main.go --output ./docs
