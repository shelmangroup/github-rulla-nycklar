FROM golang:1.13
WORKDIR /go/src/github.com/shelmangroup/github-secrets-sync

RUN go get golang.org/x/tools/cmd/goimports
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux make install
RUN CGO_ENABLED=0 GOOS=linux make test

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/bin/k8s-metadata /bin/k8s-metadata
ENTRYPOINT ["/bin/k8s-metadata"]
