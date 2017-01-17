
FROM golang:1.7.4
MAINTAINER denverdino@gmail.com

ENV OS "darwin linux windows"
ENV ARCH "amd64"

COPY . ${GOPATH}/src/github.com/denverdino/docker-machine-driver-aliyunecs
WORKDIR ${GOPATH}/src/github.com/denverdino/docker-machine-driver-aliyunecs
RUN set -ex \
	&& uname -a \
	&& go version \
	&& go env \
    && go get ./... \
    && go vet ./...
RUN dmver=$(cd $GOPATH/src/github.com/docker/machine && git describe --abbrev=0 --tags) \
    && echo "VERSION docker-machine '$dmver'"
RUN for GOOS in $OS; do \
        arch="$GOOS-$ARCH" \
        && binary="bin/docker-machine-driver-aliyunecs.$arch" \
        && echo "Building $binary" \
        && GOOS=$GOOS GOARCH=$GOARCH go build -o $binary \
        && tar czvf bin/docker-machine-driver-aliyunecs_$arch.tgz $binary; \
	done
