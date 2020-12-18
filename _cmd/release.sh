#!/bin/sh

# Environment list
# $GOOS     $GOARCH
# darwin    arm64
# darwin    amd64
# windows   386
# windows   amd64
# linux     386
# linux     amd64

# set -e

OS=("darwin" "darwin" "windows" "windows" "linux" "linux")
ARCH=("amd64" "arm64" "386" "amd64" "386" "amd64")

mkdir build
cd build

for i in `seq 0 1 5`
do
  GOOS=${OS["$i"]}
  GOARCH=${ARCH["$i"]}

  echo $GOOS $GOARCH

  EXT=""
  if [ $GOOS = "windows" ]; then
      EXT=".exe"
  fi
  OUTPUT=golin${EXT}
  GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${OUTPUT} ../golin/main.go

  gzip ${OUTPUT} -c > golin_${GOOS}_${GOARCH}.gz

  rm ${OUTPUT}
done

cd ..
