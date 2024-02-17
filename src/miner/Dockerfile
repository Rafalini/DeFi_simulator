FROM golang:1.21.1-bullseye

WORKDIR /usr/src/miner

COPY . .

RUN go mod download && go mod verify
RUN CGO_ENABLED=0 GOOS=linux go build -o /miner

EXPOSE 8080

ENTRYPOINT ["/miner"]