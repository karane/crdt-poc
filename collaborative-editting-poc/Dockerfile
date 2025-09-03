FROM golang:1.24
WORKDIR /app
COPY . .
RUN go mod init collab-crdt && go mod tidy
RUN go get github.com/gorilla/websocket
RUN go build -o app .
CMD ["./app"]
# CMD ["tail", "-f", "/dev/null"]
