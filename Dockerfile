FROM golang:1.21 AS builder
WORKDIR /app
COPY  . .
RUN go build -o proxy cmd/main.go

FROM ubuntu:23.04 as run_stage
RUN apt-get -y update && apt-get install -y tzdata -y ca-certificates && update-ca-certificates
WORKDIR /out
COPY --from=builder /app/proxy ./binary
CMD ["./binary"]