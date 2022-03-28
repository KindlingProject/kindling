#collector
#cd kindling/collector

# If developers start labeling later:
# GitCommit=$(git describe)
GitCommit=$(git rev-parse --short HEAD || echo unsupported)
echo "Git commit:" $GitCommit


docker run -it -v $PWD:/collector kindlingproject/kindling-collector-builder bash -c 'go build -o kindling-collector -ldflags="-X 'github.com/Kindling-project/kindling/collector/version.CodeVersion=$GitCommit'"'

docker build -t kindling-collector -f deploy/Dockerfile .

