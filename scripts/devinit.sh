#!/bin/bash

set -e

# Detech current os category
unameOut="$(uname -s)"
case "${unameOut}" in
    Linux*)     OS_TYPE=Linux;;
    Darwin*)    OS_TYPE=Mac;;
    CYGWIN*)    OS_TYPE=Windows;;
    MINGW*)     OS_TYPE=Windows;;
    *)          OS_TYPE="UNKNOWN:${unameOut}"
esac

echo "OS ${OS_TYPE} is deteched."
echo "Compiling ..."

# Choice file path/name by os category
if [ $OS_TYPE == "Linux" ]; then
	GOOS=linux GOARCH=amd64 go build -o bin/terraform-provider-ucloud
	chmod +x bin/terraform-provider-ucloud
    mv bin/terraform-provider-ucloud $HOME/.terraform.d/plugins
elif [ $OS_TYPE == "Mac" ]; then
	GOOS=darwin GOARCH=amd64 go build -o bin/terraform-provider-ucloud
	chmod +x bin/terraform-provider-ucloud
    mv bin/terraform-provider-ucloud $HOME/.terraform.d/plugins
elif [ $OS_TYPE == "Windows" ]; then
	GOOS=windows GOARCH=amd64 go build -o bin/terraform-provider-ucloud.exe
	chmod +x bin/terraform-provider-ucloud.exe
    mv bin/terraform-provider-ucloud.exe $APPDATA/terraform.d/plugins
else
    echo "Invalid OS"
    exit 1
fi

echo "Installation of UCloud Terraform Provider is completed."
