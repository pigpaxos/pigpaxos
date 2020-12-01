#!/usr/bin/env bash

source common.sh

arraylength=${#nodes[@]}
for (( i=1; i<${arraylength}+1; i++ ));
do
	upload_one ${nodes[$i-1]} "/home/pigpaxos/go/src/github.com/pigpaxos/pigpaxos/sigmod2021/prc_experiments/$1" "/home/ubuntu/paxi/bin/$2"
	echo "Uploaded Config to 1.$i"
done

upload_one $va_client "/home/pigpaxos/go/src/github.com/pigpaxos/pigpaxos/sigmod2021/prc_experiments/$1" "/home/ubuntu/paxi/bin/$2"
echo "Uploaded Config to Client"