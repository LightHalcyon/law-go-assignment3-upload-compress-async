ARG GO_VERSION=1.12

FROM golang:${GO_VERSION} AS builder

ENV UNAME=1406568753
ENV PW=167664
ENV URL=152.118.148.103
ENV PORT=5672
ENV VHOST=1406568753
ENV EXCNAME=1406568753

ADD . /go/src/github.com/reznov53/law-go-assignment3-upload-compress-async/server

RUN go get github.com/gin-gonic/gin
RUN go get github.com/streadway/amqp
RUN go get github.com/gin-contrib/cors
RUN go install github.com/reznov53/law-go-assignment3-upload-compress-async/server

CMD [ "/go/bin/./server" ]

EXPOSE 20606