FROM golang:1.17.3 AS builder

WORKDIR /go/src/github.com/KohlsTechnology/hierarchy
COPY . .
RUN make build

FROM scratch

COPY --from=builder /go/src/github.com/KohlsTechnology/hierarchy/hierarchy /hierarchy

ENTRYPOINT ["/hierarchy"]
