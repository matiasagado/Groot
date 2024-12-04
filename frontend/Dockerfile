FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src .
RUN  go build -o server .

EXPOSE 1323

CMD ["./server"]
