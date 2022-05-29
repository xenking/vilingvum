FROM golang:1.17 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -v -ldflags='-w -s' -o /app/build/service /app/cmd

FROM alpine
WORKDIR /app
COPY --from=build /app/build/service /usr/local/bin/service
COPY database/sql/migrations /app/database/sql/migrations
COPY static /app/static
