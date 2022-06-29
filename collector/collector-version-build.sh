GitCommit=$(git rev-parse --short HEAD || echo unsupported)
echo "Git commit:" $GitCommit
go build -o kindling-collector -ldflags="-X 'github.com/Kindling-project/kindling/collector/version.CodeVersion=$GitCommit'" ./cmd/kindling-collector/


