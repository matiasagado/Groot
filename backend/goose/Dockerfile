FROM golang:1.23-alpine AS builder

ARG GOOSE_VERSION=v3.22.1

WORKDIR /app

RUN apk add --no-cache git

RUN git clone https://github.com/pressly/goose.git \
    && cd goose \
    && git checkout $GOOSE_VERSION \
    && cd cmd/goose \
    && go build -o /goose

WORKDIR /app/goose/cmd/goose
RUN go build -o /goose

FROM alpine:3.20

WORKDIR /migrations

COPY --from=builder /goose /usr/local/bin/goose

ENTRYPOINT ["goose"]
