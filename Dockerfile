# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev

RUN CGO_ENABLED=0 GOOS=linux go build \
	-ldflags="-s -w -X main.version=${VERSION}" \
  -o /out/golt \
  ./cmd/golt

FROM alpine:3.21

LABEL org.opencontainers.image.title="Golt Runtime"
LABEL org.opencontainers.image.description="Go-native TypeScript/JavaScript backend runtime"
LABEL org.opencontainers.image.source="https://github.com/Aztekode/golt"
LABEL org.opencontainers.image.url="https://golt.dev"
LABEL org.opencontainers.image.licenses="UNLICENSED"

RUN apk add --no-cache ca-certificates \
  && addgroup -S golt \
  && adduser -S golt -G golt

COPY --from=builder /out/golt /usr/local/bin/golt

USER golt

EXPOSE 3000

ENTRYPOINT ["golt"]
CMD ["--help"]
