#!/usr/bin/env bash

source common.sh

arraylength=${#nodes[@]}
for (( i=1; i<${arraylength}+1; i++ ));
do
  upload_all ${nodes[$i-1]} "/home/pigpaxos/go/src/github.com/pigpaxos/pigpaxos/bin" "/home/ubuntu/paxi"
  echo "Uploaded to 1.$i"
done

#upload_all ${nodes[0]} "/home/pigpaxos/go/src/github.com/pigpaxos/pigpaxos/bin" "/home/ubuntu/paxi"
#echo "Uploaded to 1.1"

#upload_all $va_client "/home/pigpaxos/go/src/github.com/pigpaxos/pigpaxos/bin" "/home/ubuntu/paxi"
#echo "Uploaded to Client"
