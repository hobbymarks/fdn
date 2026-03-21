# Default recipe to run everything
all: config-go install-tools update-shell

# Configure Go environment variables globally
config-go:
    go env -w GOPROXY="https://proxy.golang.org,direct"
    go env -w GOSUMDB="sum.golang.org"
    @echo "Go Proxy and SumDB configured."

# Install essential Go development tools
install-tools:
    @echo "Installing gopls..."
    go install golang.org/x/tools/gopls@latest
    @echo "Installing goimports..."
    go install golang.org/x/tools/cmd/goimports@latest
    @echo "Installing delve (dlv)..."
    go install github.com/go-delve/delve/cmd/dlv@latest
    @echo "Installing staticcheck..."
    go install honnef.co/go/tools/cmd/staticcheck@latest

# Add Go bin to your .zshrc if it's not already there
update-shell:
    @if ! grep -q "GOPATH)/bin" ~/.zshrc; then \
        echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc; \
        echo "Added Go bin to ~/.zshrc. Please run 'source ~/.zshrc' to apply."; \
    else \
        echo "Go bin already in ~/.zshrc Path."; \
    fi

# Verify the installation
check:
    go env GOPROXY GOSUMDB
    gopls version
    dlv version
    staticcheck --version
