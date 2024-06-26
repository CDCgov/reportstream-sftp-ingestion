FROM golang:1.22 as builder

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN apt update && apt install -y --no-install-recommends ca-certificates && apt clean

RUN go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /opt/build/

COPY ./ ./

RUN make compile


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /opt/build/reportstream-sftp-ingestion /usr/local/bin/reportstream-sftp-ingestion
COPY --from=builder /go/bin/dlv /usr/local/bin/dlv
COPY --from=builder /usr/local/go/bin/go /usr/local/bin/go

# make go actually run, and we need to for dlv, and we need dlv to debug.
#RUN ["mkdir", "/tmp"]
ENV GOROOT /

CMD ["/usr/local/bin/reportstream-sftp-ingestion"]

EXPOSE 8080


# 1. Keep on trying to get mkdir working and make all the missing directories that `go` executable requires.  If there is a lot of this custom `Dockerfile` commands that we need to insert just to get this to work, that will reduce the maintainability.
# 2. Do `FROM alpine` instead of `scratch`.  This will make the docker image much larger, but have a much more "sane" Linux environment.
# 3. Make a separate `Dockerfile` for debugging and using in the `docker-compose.yml` file.  Downside of this though is it is possible for the local/debugging Dockerfile to diverge from the main/deployed `Dockerfile`.
