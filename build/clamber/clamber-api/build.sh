env GOOS=linux GOARCH=amd64 go build -v -o ./api/bin/application ../../../opensearch/cmd/api

 zip -r api.zip api

