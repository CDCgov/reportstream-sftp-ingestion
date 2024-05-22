FROM golang:1.22 as builder

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN apt update && apt install -y --no-install-recommends ca-certificates && apt clean

WORKDIR /opt/build/

COPY ./ ./

RUN make compile


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /opt/build/reportstream-sftp-ingestion /usr/local/bin/reportstream-sftp-ingestion

ENTRYPOINT ["/usr/local/bin/reportstream-sftp-ingestion"]
