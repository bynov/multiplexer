FROM golang:1.14-stretch

WORKDIR /opt/multiplexer

COPY . .

RUN go build -i ./cmd/multiplexer

FROM debian:10

RUN apt-get update && \
    apt-get install -y ca-certificates

WORKDIR /opt/multiplexer

COPY --from=0 /opt/multiplexer/multiplexer .

EXPOSE 8080

CMD [ "./multiplexer" ]