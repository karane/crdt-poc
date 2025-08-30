FROM golang:1.24
WORKDIR /app
COPY . .
RUN go mod init crdt && go mod tidy
RUN go get github.com/google/uuid
RUN go build -o app .
CMD ["./app"]
