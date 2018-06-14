FROM golang:1.10.3 as builder
RUN wget https://www.foundationdb.org/downloads/5.1.7/ubuntu/installers/foundationdb-clients_5.1.7-1_amd64.deb && dpkg -i foundationdb-clients_5.1.7-1_amd64.deb
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/github.com/bankex/go-plasma/
COPY . .
#RUN go get -d -v
RUN dep ensure -v
RUN go build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/bankex/go-plasma/go-plasma .
CMD ["./go-plasma"]