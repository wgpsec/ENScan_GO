FROM golang:1.19 AS build-stage

RUN git clone --depth 1 https://github.com/wgpsec/ENScan_GO.git

WORKDIR ENScan_GO


ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

RUN CGO_ENABLED=0 go build -trimpath -ldflags=" \
-w -s \
-X 'github.com/wgpsec/ENScan/common.BuiltAt=`date +'%F %T %z'`' \
-X 'github.com/wgpsec/ENScan/common.GoVersion=`go version | sed 's/go version //'`' \
-X 'github.com/wgpsec/ENScan/common.GitAuthor=`git show -s --format='format:%aN <%ae>' HEAD`' \
-X 'github.com/wgpsec/ENScan/common.GitCommit=`git log --pretty=format:"%h" -1`' \
-X 'github.com/wgpsec/ENScan/common.GitTag=`git describe --abbrev=0 --tags`' \
" \
-o /enscan .

# Deploy the application binary into a lean image
FROM scratch

WORKDIR /

COPY --from=build-stage /enscan /enscan

ENTRYPOINT ["/enscan"]