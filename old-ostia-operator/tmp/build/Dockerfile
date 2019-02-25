FROM alpine:3.6

ADD tmp/_output/bin/ostia-operator /usr/local/bin/ostia-operator

RUN adduser -D ostia-operator
USER ostia-operator
