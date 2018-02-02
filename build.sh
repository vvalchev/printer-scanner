#!/bin/sh
# always run from local dir
cd `dirname $0`
# fix GO environment
_GROOT=`which go`
_GROOT=`dirname $_GROOT`/..
CGO_ENABLED="0"
GOPATH=`pwd`
export CGO_ENABLED GOPATH
# make sure GO bin path exists
if [ ! -d  "$GOPATH/bin" ]; then
    mkdir "$GOPATH/bin"
fi
case "$1" in
    "build.linux")
        GOROOT="$_GROOT" go get -d -v ./...
        go build -a -ldflags "--s -extldflags '-static' -X main.Version=git:$CI_BUILD_REF" -o "printer-scanner$SUFFIX" main
        ;;
    "build.mac")
        GOOS="darwin"
        GOARCH="amd64"
        SUFFIX=".$GOOS-$GOARCH"
        export GOOS GOARCH SUFFIX
        $0 build.linux
        ;;
    "build.win")
        GOOS="windows"
        GOARCH="amd64"
        SUFFIX=".$GOOS-$GOARCH.exe"
        export GOOS GOARCH SUFFIX
        $0 build.linux
        ;;
    "build")
        $0 build.linux
        $0 build.mac
        $0 build.win
        ;;
    "shell")
        shift
        docker run -it --rm --name printer-scanner-builder -v `pwd`:/go golang:1.7 /bin/bash
        ;;
    *)
        docker run -it --rm --name printer-scanner-builder -v `pwd`:/go golang:1.7 /bin/sh -c "/go/build.sh build"
esac
