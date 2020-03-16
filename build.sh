#!/bin/sh
# always run from local dir
cd `dirname $0`
CGO_ENABLED="0"
export CGO_ENABLED
unset GOPATH # we are using go modules!
case "$1" in
    "build.linux")
        go build -a -ldflags "--s -extldflags '-static' -X main.Version=git:$CI_BUILD_REF" -o "printer-scanner$SUFFIX" ./...
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
        docker run -it --rm --name printer-scanner-builder -v `pwd`:/go golang:1.14 /bin/bash
        ;;
    *)
        docker run -it --rm --name printer-scanner-builder -v `pwd`:/go golang:1.14 /bin/sh -c "/go/build.sh build"
esac
