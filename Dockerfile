FROM golang:1.24-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -trimpath -ldflags="-s -w" -o /out/placeholder-bot .

FROM alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && \
    adduser -S -G app -h /opt/app app && \
    mkdir -p /opt/app/static && \
    chown -R app:app /opt/app

WORKDIR /opt/app
COPY --from=builder /out/placeholder-bot /opt/app/placeholder-bot

USER app
CMD ["/opt/app/placeholder-bot"]
