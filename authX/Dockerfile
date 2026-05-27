FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go mod tidy

RUN go install github.com/codegangsta/gin@latest

ENTRYPOINT ["./entrypoint.sh"]