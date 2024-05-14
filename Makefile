compile:
	go build -o ./reportstream-sftp-ingestion ./cmd/

unitTests:
	go test ./...

vet:
	go vet ./...

formatCheck:
	gofmt -l ./ && test -z "$(gofmt -l ./)"

dockerBuild:
	docker build . -t reportstream-sftp-ingestion

dockerRun:
	docker run -it reportstream-sftp-ingestion
