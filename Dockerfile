FROM golang:1.23-alpine3.21 AS builder

COPY . /github.com/KaffeeMaschina/ozon_test_task/source/
WORKDIR /github.com/KaffeeMaschina/ozon_test_task/source/

RUN go mod download
RUN go build -o ./bin/myHabr_server server.go

FROM alpine:3.21

WORKDIR /root/
COPY --from=builder /github.com/KaffeeMaschina/ozon_test_task/source/bin/myHabr_server .

CMD ["./myHabr_server", "-usePostgres"]