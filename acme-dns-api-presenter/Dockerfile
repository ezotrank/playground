FROM golang:1.17-alpine as backend

ENV CGO_ENABLED=0

ADD . /build
WORKDIR /build

RUN apk add --no-cache --update tzdata ca-certificates
RUN go build -o /build/acme-dns-api-presenter

FROM scratch

COPY --from=backend /build/acme-dns-api-presenter /srv/acme-dns-api-presenter
COPY --from=backend /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /srv
ENTRYPOINT ["/srv/acme-dns-api-presenter"]
