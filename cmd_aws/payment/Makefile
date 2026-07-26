.PHONY: zip build test clean

GOOS=linux
GOARCH=arm64
CGO_ENABLED=0

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -tags lambda.norpc -o bootstrap .

zip: build
	zip -FS myFunction.zip bootstrap countries_iso_slim.json

test:
	go test ./...

clean:
	rm -f bootstrap payment myFunction.zip
