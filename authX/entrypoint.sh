#!/bin/sh

if [ "$DEBUG" = "1" ]; then 
	RED='\033[0;31m'
	NC='\033[0m' # No Color
	echo "${RED}DEBUG mode is ON${NC}"
fi

cd /app/cmd/

# export GIN_BUILD_ARGS="-race"

gin -i main.go