FROM golang:alpine

MAINTAINER Kristoph Junge <kristoph.junge@gmail.com>

WORKDIR /go

COPY . .

RUN go build -v -o bin/app src/app.go

CMD ["./bin/app"]
