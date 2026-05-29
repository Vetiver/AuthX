FROM golang:1.25.0

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go mod tidy

RUN go install github.com/codegangsta/gin@latest

COPY . .

RUN chmod +x entrypoint.sh

ENV GOOS=linux
ENV GOARCH=amd64

ENTRYPOINT ["./entrypoint.sh"]