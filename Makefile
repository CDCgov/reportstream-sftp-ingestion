compile:
	go build -o ./reportstream-sftp-ingestion ./cmd/

unitTests:
	go test ./...

vet:
	go vet ./...

formatCheck:
	gofmt -l ./ && test -z "$(gofmt -l ./)"
