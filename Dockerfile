FROM golang:1.10.3 as builder
RUN wget https://www.foundationdb.org/downloads/5.2.5/ubuntu/installers/foundationdb-clients_5.2.5-1_amd64.deb && dpkg -i foundationdb-clients_5.2.5-1_amd64.deb
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/github.com/shamatar/go-plasma/
COPY . .
COPY fdb.cluster /etc/foundationdb/fdb.cluster
EXPOSE 3001
RUN dep ensure -v
RUN cd ../../matterinc/PlasmaCommons
RUN git submodule init
RUN git submodule update --recursive
#RUN cd crypto/secp256k1/
#RUN git clone https://github.com/bitcoin-core/secp256k1.git
RUN cd ../..
CMD ["go", "run", "server.go"]
# CMD ["go test -v loadTest/createAndSpend_test.go"]