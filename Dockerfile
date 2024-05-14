FROM golang:1.22 as builder

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN apt update

WORKDIR /opt/build/

COPY ./ ./

RUN make compile


FROM scratch

COPY --from=builder /opt/build/reportstream-sftp-ingestion /usr/local/bin/reportstream-sftp-ingestion

ENTRYPOINT ["/usr/local/bin/reportstream-sftp-ingestion"]
