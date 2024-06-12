FROM golang:1.20 AS base

FROM base AS builder

WORKDIR /build
COPY . ./
RUN go build ./cmd/api

FROM base AS final

ARG PORT

WORKDIR /app
COPY --from=builder /build/api /build/.env ./

EXPOSE ${PORT}
CMD ["/app/api"]