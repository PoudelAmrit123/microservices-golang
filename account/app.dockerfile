FROM golang:1.22-alpine AS build
RUN apk --no-cache add gcc g++ make ca-certificates
WORKDIR /go/src/github.com/PoudelAmrit123/microservice
# WORKDIR /app

COPY go.mod go.sum ./
COPY vendor vendor
COPY account account

# RUN mkdir -p /go/bin && chmod -R 777 /go/bin
RUN mkdir -p /app && chmod -R 777 /app
RUN   ls -ld /go/src
# RUN GO111MODULE=on go build -mod vendor -o /go/bin/app ./account/cmd/account
RUN GO111MODULE=on go build -o /app/app ./account/cmd/account


FROM alpine:3.11
WORKDIR /usr/bin
# COPY --from=build /go/bin .
COPY --from=build /app .
EXPOSE 8080
CMD ["app"]