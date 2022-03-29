
cd ..
#/kindling

docker run -it -v $PWD:/kindling kindlingproject/kindling-collector-builder bash

cd ..
cd kindling/collector
#/kindling/collector


GitCommit=$(git rev-parse --short HEAD || echo unsupported)
echo "Git commit:" $GitCommit
go build -o kindling-collector -ldflags="-X 'github.com/Kindling-project/kindling/collector/version.CodeVersion=$GitCommit'"

