ARG GO_VER=1.24.2
ARG ALPINE_VER=3.20

FROM golang:${GO_VER}-alpine${ALPINE_VER}

# Install air for hot reloading
RUN go install github.com/air-verse/air@latest

WORKDIR /go/src/github.com/chainlaunch/chaincode-fabric-go-tmpl
COPY ./go.mod ./go.sum ./
COPY . .

RUN go mod download

EXPOSE 9999
CMD ["air"]
