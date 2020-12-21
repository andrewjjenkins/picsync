FROM golang:latest as build

RUN mkdir -p /go/src/github.com/andrewjjenkins/picsync
WORKDIR /go/src/github.com/andrewjjenkins/picsync
COPY go.mod go.sum ./
RUN go mod download

COPY . /go/src/github.com/andrewjjenkins/picsync
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a --installsuffix nocgo -o /picsync ./cmd/picsync
CMD ["/picsync"]

FROM alpine:3.8 as run
RUN apk add --no-cache ca-certificates
COPY --from=build /picsync ./
CMD ["./picsync"]
