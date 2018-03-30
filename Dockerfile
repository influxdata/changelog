FROM golang:1.9.3 as builder
COPY . /go/src/github.com/influxdata/changelog
RUN go get github.com/golang/dep/cmd/dep && \
    cd /go/src/github.com/influxdata/changelog && \
    dep ensure -vendor-only && \
    go install ./cmd/git-changelog

FROM buildpack-deps:stretch-scm
COPY --from=builder /go/bin/git-changelog /usr/bin/git-changelog
