# syntax=docker/dockerfile:1
FROM golang:1.19-alpine AS builder

ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV GOCACHE=/go-build

WORKDIR /code

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod/cache \
    go mod download

RUN mkdir /build

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/cache \
    --mount=type=cache,target=/go-build \
    go build -o /build/certs cmd/main.go

CMD [ "/build/certs" ]

FROM alpine AS release
WORKDIR /out
COPY --from=builder /build/certs /usr/local/bin/certs
CMD ["/usr/local/bin/certs"]