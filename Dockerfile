FROM golang:1.14-alpine

RUN apk add --no-cache git curl
RUN mkdir -p /go/src/github.com/theblueskies/spothro
WORKDIR /go/src/github.com/theblueskies/spothro

COPY . ./
EXPOSE 9000

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o  rate-service .

CMD ["./rate-service"]
