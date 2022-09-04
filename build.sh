#!/bin/bash

appName="enscan"
builtAt="$(date +'%F %T %z')"
goVersion=$(go version | sed 's/go version //')
gitAuthor=$(git show -s --format='format:%aN <%ae>' HEAD)
gitCommit=$(git log --pretty=format:"%h" -1)

if [ "$1" == "release" ]; then
  gitTag=$(git describe --abbrev=0 --tags)
else
  gitTag=build-next
fi

echo "build version: $gitTag"

ldflags="\
-w -s \
-X 'github.com/wgpsec/ENScan/common.BuiltAt=$builtAt' \
-X 'github.com/wgpsec/ENScan/common.GoVersion=$goVersion' \
-X 'github.com/wgpsec/ENScan/common.GitAuthor=$gitAuthor' \
-X 'github.com/wgpsec/ENScan/common.GitCommit=$gitCommit' \
-X 'github.com/wgpsec/ENScan/common.GitTag=$gitTag' \
"

if [ "$1" == "release" ]; then
  xgo -out enscan -ldflags="$ldflags" .
else
  xgo -targets=linux/amd64,windows/amd64,darwin/amd64 -out enscan -ldflags="$ldflags"  .
fi

mkdir "build"
mv enscan-* build
cd build || exit
upx -9 ./*
find . -type f -print0 | xargs -0 md5sum > md5.txt
cat md5.txt
# compress file (release)
if [ "$1" == "release" ]; then
    mkdir compress
    mv md5.txt compress
    for i in `find . -type f -name "$appName-linux-*"`
    do
      tar -czvf compress/"$i".tar.gz "$i"
    done
    for i in `find . -type f -name "$appName-darwin-*"`
    do
      tar -czvf compress/"$i".tar.gz "$i"
    done
    for i in `find . -type f -name "$appName-windows-*"`
    do
      zip compress/$(echo $i | sed 's/\.[^.]*$//').zip "$i"
    done
fi
cd ../..