FROM golang:1.17-alpine3.15 as build

ENV GOFLAGS="-mod=vendor" TZ="Europe/Kiev"
WORKDIR /build
ADD . /build
RUN \
    apk add --no-cache git && \
    version=$(git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD) && \
    commit=$(git rev-parse --short HEAD) && \
    date=$(date -u +"%Y-%m-%dT%H:%M:%SZ") && \
    go version && \
    CGO_ENABLED=0 go build -o p24 -ldflags "-X main.version=${version} -X main.commit=${commit} -X main.date=${date} -s -w" .

FROM alpine:3.15

ENV \
    TERM=xterm-color \
    TZ=Europe/Kiev   \
    APP_USER=app     \
    APP_UID=1000

COPY ./entrypoint.sh /entrypoint.sh
COPY --from=build /build/p24 /usr/local/bin/p24

RUN \
    apk add --no-cache --update su-exec tzdata ca-certificates dumb-init && rm -rf /var/cache/apk/* && \
    adduser -s /bin/sh -D -u $APP_UID $APP_USER && \
    mkdir -p /app && chown -R $APP_USER:$APP_USER /app && \
    chmod +x /entrypoint.sh /usr/local/bin/p24

WORKDIR /app
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/local/bin/p24"]
