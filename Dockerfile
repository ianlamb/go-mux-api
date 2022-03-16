FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o /go-mux-api

# RUN apk add --no-cache ca-certificates && update-ca-certificates

EXPOSE 8010 8010

CMD ["/go-mux-api"]
