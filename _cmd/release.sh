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
Version="v2.0.0beta"
Revision==$(git rev-parse --short HEAD)

echo "============================================"
echo "Build Version: ${Version}"
echo "-Git revision: ${Revision}"

OS=("darwin" "darwin" "windows" "windows" "linux" "linux")
ARCH=("amd64" "arm64" "386" "amd64" "386" "amd64")

rm -r build
mkdir build
cd build

for i in `seq 0 1 5`
do
  GOOS=${OS["$i"]}
  GOARCH=${ARCH["$i"]}

  echo "Build OS=$GOOS ARCHITECT=$GOARCH start"

  EXT=""
  if [ $GOOS = "windows" ]; then
      EXT=".exe"
  fi
  OUTPUT=golin${EXT}
  GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "-X main.version=${Version} -X main.revision=${Revision}" -o ${OUTPUT} ../golin/main.go

  gzip -9 ${OUTPUT} -c > golin_${GOOS}_${GOARCH}.gz

  rm ${OUTPUT}
done

cd ..

echo "Success"
echo "============================================"
