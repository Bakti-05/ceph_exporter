FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /ceph_exporter

FROM alpine:3.9

COPY --from=builder /ceph_exporter /ceph_exporter

EXPOSE 9128

ENTRYPOINT ["/ceph_exporter"]
