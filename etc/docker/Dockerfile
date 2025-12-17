FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY cmd ./cmd
COPY proto ./proto
COPY src/backend ./src/backend

COPY buf.yaml .
COPY Makefile .

COPY etc/scripts/tool-install.sh .

RUN chmod +x tool-install.sh && ./tool-install.sh

RUN make buf-gen-go

RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/run.go

FROM alpine:latest

ENV PB_SERVER_PORT=8000
ENV PB_SERVER_HOST=0.0.0.0

WORKDIR /app

COPY --from=builder /app/server .
COPY etc/scripts/server-entrypoint.sh .

RUN chmod +x server-entrypoint.sh

CMD ["./server-entrypoint.sh"]
