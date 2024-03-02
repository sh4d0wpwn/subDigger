#!/bin/bash

sudo apt update
sudo apt install -y python3-pip
sudo apt install unzip

# Install Sublist3r
pip3 install Sublist3r

# Install Subfinder
go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest

# Install Assetfinder
go install -v github.com/tomnomnom/assetfinder@latest

# Install Findomain
FINDOMAIN_VERSION="9.0.4" # Replace with the latest version
wget "https://github.com/Findomain/Findomain/releases/download/${FINDOMAIN_VERSION}/findomain-linux.zip"
unzip "findomain-linux.zip" -d findomain
chmod +x findomain/findomain
sudo mv findomain/findomain /usr/local/bin/
sudo chown ubuntu:ubuntu /usr/local/bin/findomain
rm -r findomain "findomain-linux.zip"

# Verifying installations
echo "Verifying installations..."
go version
sublist3r -h
subfinder -h
assetfinder -h
findomain --version
echo "Installation of tools complete."
