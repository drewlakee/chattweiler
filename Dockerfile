FROM golang:1.17

WORKDIR /application

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY ./ ./
RUN go build -o ./chattweiler ./cmd/main.go

CMD ["./chattweiler"]