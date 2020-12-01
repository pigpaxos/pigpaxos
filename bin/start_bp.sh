#!/usr/bin/env bash

PID_FILE=server.pid

./server -log_dir=logs -log_level=info -algorithm=pigpaxos -id=$1 -pg=$2 -rgslack=$3 -fr=$4 -stdpigtimeout=$5> logs/out1.$1.txt 2>&1 &
echo $! >> ${PID_FILE}
