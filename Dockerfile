FROM golang:1.9.3 as builder
COPY . /go/src/github.com/influxdata/changelog
RUN go get -d github.com/influxdata/changelog/... && \
    go install github.com/influxdata/changelog/...

FROM buildpack-deps:stretch-scm
COPY --from=builder /go/bin/git-changelog /usr/bin/git-changelog
