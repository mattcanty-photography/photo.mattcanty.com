build:
	rm -f .build/www.zip
	GOOS=linux GOARCH=amd64 go build -o ./.build/www ./main.go
	zip -j ./.build/www.zip ./.build/www *.html
