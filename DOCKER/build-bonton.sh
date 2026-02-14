#!/usr/bin/env bash
set -e

# Get the tag from the version, or try to figure it out.
if [ -z "$TAG" ]; then
	TAG=$(awk -F\" '/TMCoreSemVer =/ { print $1; exit }' < ../version/version.go)
fi
if [ -z "$TAG" ]; then
		echo "Please specify a tag."
		exit 1
fi

read -p "==> Build docker images with the following tag $TAG? y/n" -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]
then
		docker build -t "aleksriaden/cometbft:$TAG" --file=Dockerfile.bonton .
fi
