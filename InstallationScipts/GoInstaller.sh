#!/bin/bash

# Set the Go version
GO_VERSION=1.22.0
GO_TAR_FILE=go$GO_VERSION.linux-amd64.tar.gz

# Update and upgrade system packages
echo "Updating and upgrading system packages..."
sudo apt update && sudo apt upgrade -y

# Download the specified Go version
echo "Downloading Go $GO_VERSION..."
wget https://golang.org/dl/$GO_TAR_FILE

# Extract the Go archive
echo "Extracting Go archive to /usr/local..."
sudo tar -C /usr/local -xzf $GO_TAR_FILE

# Set up Go environment variables
echo "Setting up Go environment variables..."
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.profile
echo "export GOPATH=\$HOME/go" >> ~/.profile
echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.profile
source ~/.profile

# Create Go workspace directories
echo "Creating Go workspace directories..."
mkdir -p $HOME/go/{bin,src,pkg}

# Clean up the downloaded tar file
echo "Cleaning up..."
rm $GO_TAR_FILE

# Verify Go installation
echo "Verifying Go installation..."
go version

echo "Go installation is complete."
