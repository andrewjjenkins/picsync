FROM golang:latest as build

RUN mkdir -p /go/src/github.com/andrewjjenkins/picsync
WORKDIR /go/src/github.com/andrewjjenkins/picsync
COPY Gopkg.lock Gopkg.toml ./
RUN go get -u github.com/golang/dep/... && \
  dep ensure -vendor-only

COPY . /go/src/github.com/andrewjjenkins/picsync
RUN pwd && ls -lR pkg/ cmd/
RUN CGO_ENABLED=0 GOOS=linux go build -a --installsuffix nocgo -o /picsync ./cmd/picsync
CMD ["/picsync"]

FROM scratch as run
COPY --from=build /picsync ./
CMD ["./picsync"]
