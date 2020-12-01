#!/usr/bin/env bash

PID_FILE=server.pid

PID=$(cat "${PID_FILE}");

if [ -z "${PID}" ]; then
    echo "Process id for servers is written to location: {$PID_FILE}"
    go build ../server/
    go build ../client/
    go build ../cmd/
    rm -r logs
    mkdir logs/
    ./server -log_dir=logs -log_level=debug -id 1.1 -algorithm=pigpaxos -pg=1 -rgslack=4>logs/out1.1.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=debug -id 1.2 -algorithm=pigpaxos -pg=1 -rgslack=4>logs/out1.2.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=debug -id 1.3 -algorithm=pigpaxos -pg=1 -rgslack=4>logs/out1.3.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=debug -id 1.4 -algorithm=pigpaxos -pg=1 -rgslack=4>logs/out1.4.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=debug -id 1.5 -algorithm=pigpaxos -pg=1 -rgslack=4>logs/out1.5.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=debug -id 1.6 -algorithm=pigpaxos -pg=1 -rgslack=4>logs/out1.6.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=debug -id 1.7 -algorithm=pigpaxos -pg=1 -rgslack=4>logs/out1.7.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=debug -id 1.8 -algorithm=pigpaxos -pg=1 -rgslack=4>logs/out1.8.txt 2>&1 &
    echo $! >> ${PID_FILE}
    ./server -log_dir=logs -log_level=debug -id 1.9 -algorithm=pigpaxos -pg=1 -rgslack=4>logs/out1.9.txt 2>&1 &
    echo $! >> ${PID_FILE}

    sleep 2
    echo "starting clients"
    #./client -id=1.1
    ./client &

    #sleep 2
    #crashpid=`head -n 1 ${PID_FILE}`
    #echo "Crashing ${crashpid}"
    #kill -15 "${crashpid}"


else
    echo "Servers are already started in this folder."
    exit 0
fi

