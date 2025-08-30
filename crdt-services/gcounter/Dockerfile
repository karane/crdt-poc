FROM golang:1.24
WORKDIR /app
COPY . .
RUN go mod init crdt && go mod tidy
RUN go build -o app .
CMD ["./app"]
# CMD ["tail", "-f", "/dev/null"]
