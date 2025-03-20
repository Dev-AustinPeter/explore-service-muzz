FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

WORKDIR /app/services/explore

RUN go build -o explore_service

EXPOSE 50051

CMD ["/app/services/explore/explore_service"]
