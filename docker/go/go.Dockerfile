FROM golang:1.21 AS base

FROM base AS builder

WORKDIR /build
COPY . ./
RUN go build ./cmd/api

FROM base AS final

ARG PORT

WORKDIR /app
COPY --from=builder /build/api /build/.env /build/keys/gmail_key.json ./

EXPOSE ${PORT}
CMD ["/app/api"]