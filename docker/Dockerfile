FROM golang

ADD  . /go/src/github.com/littlebunch/gnutdata-bfpd-api/
WORKDIR /go/src/github.com/littlebunch/gnutdata-bfpd-api/

RUN go get github.com/fvbock/endless
RUN go get github.com/gin-gonic/gin
RUN go get gopkg.in/couchbase/gocb.v1
#RUN go get github.com/littlebunch/gnutdata-bfpd-api/model
RUN go get gopkg.in/couchbase/gocb.v1/cbft
RUN go build -o /go/bin/bfpd github.com/littlebunch/gnutdata-bfpd-api/api/

ENTRYPOINT /go/bin/bfpd

EXPOSE 8000

