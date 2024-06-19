compile:
	cd ./src/ && go build -o ../reportstream-sftp-ingestion ./cmd/

unitTests:
	cd ./src/ && go test ./... -cover -coverprofile=coverage.out

unitTestsWithCoverageThreshold: unitTests
	./codeCoverageFilter.sh
	./codeCoverageThresholdCheck.sh

vet:
	cd ./src/ && go vet ./...

formatCheck:
	cd ./src/ && gofmt -l ./ && test -z "$(gofmt -l ./)"

formatApply:
	cd ./src/ && gofmt -w ./

dockerBuild:
	docker build . -t reportstream-sftp-ingestion

dockerRun:
	docker run -it reportstream-sftp-ingestion

dockerComposeRun:
	docker compose up --build
