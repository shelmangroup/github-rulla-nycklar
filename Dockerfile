FROM golang:1.14
WORKDIR /go/src/github.com/shelmangroup/github-rulla-nycklar

RUN go get golang.org/x/tools/cmd/goimports
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux make install
RUN CGO_ENABLED=0 GOOS=linux make test

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/bin/github-rulla-nycklar /bin/github-rulla-nycklar
ENTRYPOINT ["/bin/github-rulla-nycklar"]
