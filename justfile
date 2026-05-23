# Build the fdn binary in the current directory (version from git tag)
build:
    go build -ldflags "-X github.com/hobbymarks/fdn/cmd.version=$(git describe --tags --always --dirty)" -o fdn .

# Run all tests
test:
    go test ./...

# Run all tests with verbose output (no cache)
test-verbose:
    go test -v -count=1 ./...

# Format all Go source files
fmt:
    gofmt -w .

# Regenerate db/default_emoji_sepwords.txt from Unicode emoji data files
emoji-sepwords:
    cd db && go run ./emojigen -dir . -o default_emoji_sepwords.txt
