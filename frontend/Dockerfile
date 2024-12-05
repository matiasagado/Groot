FROM golang:1.23-alpine AS build

WORKDIR /app

RUN apk add --no-cache build-base

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src .
RUN CGO_ENABLED=1 go build -o server .

FROM golang:1.23-alpine

WORKDIR /app

COPY --from=build /app/server .

COPY src .

EXPOSE 1323

CMD ["./server"]
