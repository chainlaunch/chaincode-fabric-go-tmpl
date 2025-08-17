ARG GO_VER=1.24.2

FROM golang:${GO_VER}

# Install git and air for hot reloading
RUN apt-get update && apt-get install -y git && \
    go install github.com/air-verse/air@latest

WORKDIR /go/src/github.com/chainlaunch/chaincode-fabric-go-tmpl
COPY ./go.mod ./go.sum ./
COPY . .

RUN go mod download

EXPOSE 9999
CMD ["air"]
