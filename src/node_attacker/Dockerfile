FROM golang:1.21.1-bullseye

WORKDIR /usr/src/node

COPY . .

RUN go mod download && go mod verify
RUN CGO_ENABLED=0 GOOS=linux go build -o /node

EXPOSE 8080

ENTRYPOINT ["/node"]