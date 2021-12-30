FROM golang:1.17-alpine as backend

ENV CGO_ENABLED=0

COPY go.mod go.sum /build/
WORKDIR /build

RUN go mod download

COPY . /build

RUN go build -o /build/out/wallet /build/wallet
RUN go build -o /build/out/bank /build/bank

FROM scratch as wallet

COPY --from=backend /build/out/wallet /srv/wallet

WORKDIR /srv
ENTRYPOINT ["/srv/wallet"]

FROM scratch as bank

COPY --from=backend /build/out/bank /srv/bank

WORKDIR /srv
ENTRYPOINT ["/srv/bank"]