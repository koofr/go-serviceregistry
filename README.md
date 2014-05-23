go-serviceregistry
==================

Go ZooKeeper service registry.

[![GoDoc](https://godoc.org/github.com/koofr/go-serviceregistry?status.png)](https://godoc.org/github.com/koofr/go-serviceregistry)

## Install

    go get github.com/koofr/go-serviceregistry

## Example

    go get github.com/koofr/go-zkutils/zkserver
    ZKROOT=/opt/zookeeper-3.4.6 zkserver

    go get github.com/koofr/go-serviceregistry/zkregistryclient

    zkregistryclient register myservice http localhost:8000
    zkregistryclient get myservice http

## Testing

    go get -t
    ZKROOT=/opt/zookeeper-3.4.6 go test
