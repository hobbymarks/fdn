# Build the fdn binary in the current directory
build:
    go build -o fdn .

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
