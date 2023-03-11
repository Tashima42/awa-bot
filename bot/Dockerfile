FROM golang:1.19.1-alpine

WORKDIR /app
COPY ./bot/go.mod ./
COPY ./bot/go.sum ./
RUN go mod download
COPY ./bot ./
RUN go build -o awabot.out
CMD ./awabot.out ${ARGS}