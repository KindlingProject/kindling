cd ../../
mkdir -p probe/build
cd probe/build
cmake -DBUILD_DRIVER=OFF -DPROBE_VERSION=0.1.1dev ..
make
libKindlingPath="./src/libkindling.so"
if [ ! -f "$libKindlingPath" ]; then
  echo "compiler libkindling failed! exit!"

else
  cp -rf ./src/libkindling.so ../../collector/docker/libso
  cp -rf ./src/libkindling.so /usr/lib64/
  cd ../../collector/
  go build
  collectorPath="./collector"
  if [ ! -f "$collectorPath" ]; then
    echo "compiler collector failed! exit!"
  else
    cp -rf ./collector ./docker/kindling-collector
    cd docker
    docker build -t kindling-collector .
  fi
fi