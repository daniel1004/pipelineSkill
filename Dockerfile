FROM golang:latest AS build
WORKDIR /go/src/pipeline
RUN go env -w GO111MODULE=auto
ADD main.go .
RUN go install ./...

FROM alpine:latest
LABEL version="1.0.0"
LABEL maintainer="Daniel Kreider<kreiderdaniel10@gmail.com>"
WORKDIR /root/
COPY --from=build /go/bin/pipeline .
ENTRYPOINT ["./pipeline"]