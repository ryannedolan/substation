FROM golang:1.14 AS builder

WORKDIR /go/src
COPY . .
RUN go mod download
RUN go install ./...

FROM gcr.io/distroless/base-debian10
COPY --from=builder /go/bin/ /bin
COPY config /etc/
ENTRYPOINT ["/bin/help"]
