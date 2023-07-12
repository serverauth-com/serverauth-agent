# /bin/bash

if [ ! -t 0 ]; then x-terminal-emulator -e "$0"; exit 0; fi
echo "Before proceeding please ensure you have installed Golang. In addition this file should be run from the top level of the git repository."
read -n1 -r -p "Press any key to start build process..." key
echo "What version do you want to tag this as? Please use the format of x.x.x (e.g 1.1.0)"
read version
echo "Downloading required modules"
go mod download


go get github.com/fatih/color
go get github.com/spf13/cobra
go get github.com/spf13/viper

echo "Building Linux 64bit package"
env GOOS=linux GOARCH=amd64 go build -o serverauth ./main.go
echo "Creating archive of Linux 64bit package"
tar zcvf serverauth-agent-v"$version"-linux-64bit.tar.gz ./serverauth

echo "Building Linux 32bit package"
env GOOS=linux GOARCH=386 go build -o serverauth ./main.go
echo "Creating archive of Linux 32bit package"
tar zcvf serverauth-agent-v"$version"-linux-32bit.tar.gz ./serverauth

echo "Building Linux ARMv5 package"
env GOOS=linux GOARCH=arm GOARM=5 go build -o serverauth ./main.go
echo "Creating archive of Linux ARMv5 package"
tar zcvf serverauth-agent-v"$version"-linux-armv5.tar.gz ./serverauth

echo "Building Linux ARMv6 package"
env GOOS=linux GOARCH=arm GOARM=6 go build -o serverauth ./main.go
echo "Creating archive of Linux ARMv6 package"
tar zcvf serverauth-agent-v"$version"-linux-armv6.tar.gz ./serverauth

echo "Building Linux ARMv7 package"
env GOOS=linux GOARCH=arm GOARM=7 go build -o serverauth ./main.go
echo "Creating archive of Linux ARMv7 package"
tar zcvf serverauth-agent-v"$version"-linux-armv6.tar.gz ./serverauth


echo "The ServerAuth release files have been created. These can now be added to the release"
