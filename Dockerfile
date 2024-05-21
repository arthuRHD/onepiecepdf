FROM golang:1.22 as builder

ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o ./server ./...

FROM scratch
COPY lelscans.crt /etc/ssl/certs/lelscans.crt
COPY --from=builder /app/server /server
COPY --from=builder /app/static /static
ENTRYPOINT ["/server"]
