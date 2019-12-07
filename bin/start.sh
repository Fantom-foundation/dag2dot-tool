#!/bin/bash

# This script will start $N$ instances of dot-tools
# each of the dot-tool connects to a running node
#
# Example usage:
#
# Start capturing 4 nodes using dot-tool in root mode
# ./bin/start.sh 4 root
#
# Start capturing 5 nodes using dot-tool in epoch mode
# ./bin/start.sh 5 epoch
set -e

# number of nodes N
N=$1

#
EXEC=./bin/dot-tool
mode=$2

# default ip using localhost
IP=127.0.0.1
# default port PORT
# the actual ports are PORT+1, PORT+2, etc (18541, 18542, 18543, ... )
#PORT=18540
PORT=4000


declare -r DOT_DIR=./lachesis_images

mkdir -p ${DOT_DIR}


echo -e "\nStarting dot-tool:\n"
for i in $(seq $N)
do
    port=$((PORT + i))
    d=${DOT_DIR}/${i}
    mkdir -p $d

    echo " starting at port: ${port}, image folder: ${d}"
    ${EXEC} -mode ${mode} -host localhost \
	-port ${port} -out ${d} >${d}.log 2>${d}.err &
done

