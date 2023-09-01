cd ../../
git submodule update --init
mkdir -p src/probe/build
cd src/probe/build
agent_libs_src="/camera-agent/src/probe/libs/agent-libs/"
cflag="-DBUILD_DRIVER=OFF"
if [ -d "$agent_libs_src" ]; then
  cflag=$cflag" -DAGENT_LIBS_SOURCE_DIR=$agent_libs_src"
fi
cmake $cflag ..
make
libKindlingPath="./src/libkindling.so"
if [ ! -f "$libKindlingPath" ]; then
  echo "compiler libkindling failed! exit!"

else
  mkdir -p ../../../docker/libso && cp -rf ./src/libkindling.so ../../../docker/libso/
  cp -rf ./src/libkindling.so /usr/lib64/
  cd ../../../
  cargo build --release
  agentPath="./target/release/camera-agent"
  if [ ! -f "$agentPath" ]; then
    echo "compiler camera-agent failed! exit!"
  else
    cp ./target/release/camera-agent docker/
    cd docker
    docker build -t camera-agent . ;

  fi
fi
