FROM golang:1.23 as builder

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN apt update && apt upgrade && apt clean

RUN useradd myLowPrivilegeUser
USER myLowPrivilegeUser

WORKDIR /home/myLowPrivilegeUser/app/

COPY --chown=myLowPrivilegeUser ./ ./

RUN make compile


FROM alpine

RUN apk update && apk upgrade && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=builder /home/myLowPriviledgeUser/app/reportstream-sftp-ingestion /usr/local/bin/reportstream-sftp-ingestion

RUN adduser -S myLowPrivilegeUser
USER myLowPrivilegeUser

ENTRYPOINT ["/usr/local/bin/reportstream-sftp-ingestion"]

EXPOSE 8080
