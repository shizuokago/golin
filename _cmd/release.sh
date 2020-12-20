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
Version="v2.0.0rc1"
Revision=$(git rev-parse --short HEAD)
Date=$(date -u -R)

echo "============================================"
echo "Build Version: ${Version}"
echo "-Git revision: ${Revision}"

OS=("darwin" "darwin" "windows" "windows" "linux" "linux")
ARCH=("amd64" "arm64" "386" "amd64" "386" "amd64")

rm -r build
mkdir build

cd build

cp ../../README.md ./

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

  GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "-X 'main.version=${Version}' -X 'main.revision=${Revision}' -X 'main.date=${Date}'" -o ${OUTPUT} ../golin/main.go

  ZIPNAME="golin_${GOOS}_${GOARCH}.zip"

  echo "Compress"
  go run ../golin/main.go compress ${ZIPNAME} $OUTPUT

  rm ${OUTPUT}
done

rm README.md
cd ..

echo "Success"
echo "============================================"
