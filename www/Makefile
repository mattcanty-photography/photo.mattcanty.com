build:
	rm -f .build/www.zip
	GOOS=linux GOARCH=amd64 go build -o ./.build/www ./*.go
	zip -j ./.build/www.zip ./.build/www *.html

run:
	go run *.go
