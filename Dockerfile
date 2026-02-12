FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod vendor && \
    cd vendor/github.com/usenwep/nwep-go && bash setup.sh
RUN go build -mod=vendor -o server .

FROM debian:trixie-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 6937/udp
CMD ["./server"]
