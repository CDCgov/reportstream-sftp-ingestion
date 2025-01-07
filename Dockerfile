FROM golang:1.23 as builder

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN apt-get update -y && apt-get upgrade -y && apt-get clean -y

RUN useradd myLowPrivilegeUser
USER myLowPrivilegeUser

WORKDIR /home/myLowPrivilegeUser/app/

COPY --chown=myLowPrivilegeUser ./ ./

RUN make compile


FROM alpine:3.21.1


RUN apk update && apk upgrade && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY --from=builder /home/myLowPrivilegeUser/app/reportstream-sftp-ingestion /usr/local/bin/reportstream-sftp-ingestion

RUN adduser -S myLowPrivilegeUser
USER myLowPrivilegeUser

WORKDIR /home/myLowPrivilegeUser/

ENTRYPOINT ["/usr/local/bin/reportstream-sftp-ingestion"]

EXPOSE 8080
