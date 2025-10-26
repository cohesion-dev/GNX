FROM golang:1.24-alpine

RUN set -eux && \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk --no-cache add ca-certificates tzdata curl bash htop vim less wget lsof inetutils-telnet jq tmux zstd git

RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct && go install github.com/cespare/reflex@latest 
ENV PATH="/go/bin:${PATH}"

WORKDIR /app/backend
COPY entrypoint.sh /entrypoint.sh
RUN chmod a+x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]