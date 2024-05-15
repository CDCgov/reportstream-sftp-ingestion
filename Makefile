compile:
	go build -o ./reportstream-sftp-ingestion ./cmd/

unitTests:
	go test ./...

vet:
	go vet ./...

formatCheck:
	gofmt -l ./ && test -z "$(gofmt -l ./)"

formatApply:
	gofmt -w ./

dockerBuild:
	docker build . -t reportstream-sftp-ingestion

dockerRun:
	docker run -it reportstream-sftp-ingestion
