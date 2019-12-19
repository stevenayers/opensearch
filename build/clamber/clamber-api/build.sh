env GOOS=linux GOARCH=amd64 go build -v -o ./api/bin/application ../../../clamber/cmd/api

 zip -r api.zip api

