#collector
#cd kindling/collector

#We haven't posted any tags,so i use commitId instead
# GitCommit=$(git describe)
GitCommit=$(git rev-parse --short HEAD || echo unsupported)
echo "Git commit:" $GitCommit


# docker run -it -v $PWD:/collector kindlingproject/kindling-collector-builder bash 


# command in formal kindlingproject if merged:
# docker run -it -v $PWD:/collector kindlingproject/kindling-collector-builder bash -c 'go build -o kindling-collector -ldflags="-X 'github.com/Kindling-project/kindling/collector/version.CodeVersion=$GitCommit'"'
docker run -it -v $PWD:/collector kindlingproject/kindling-collector-builder bash -c 'go build -o kindling-collector -ldflags="-X 'github.com/sugary199/collector-version/version.CodeVersion=$GitCommit'"'

docker build -t kindling-collector -f deploy/Dockerfile .

