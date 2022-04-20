FROM golang:1.17-alpine as builder
LABEL stage=builder
WORKDIR /usr/src/app
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
 apk add --no-cache upx ca-certificates tzdata
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o bot . &&\
 upx --best bot -o _upx_server && \
 mv -f _upx_server bot

FROM scratch as runner
COPY --from=builder /usr/share/zoneinfo/UTC /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/src/app/bot /opt/app/
CMD ["/opt/app/bot"]