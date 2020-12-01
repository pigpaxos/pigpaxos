#!/usr/bin/env bash

source common.sh

arraylength=${#nodes[@]}
for (( i=1; i<${arraylength}+1; i++ ));
do
  ssh -i va.pem ubuntu@${nodes[$i-1]} 'cd /home/ubuntu/paxi/bin/; rm logs/*'
  upload_one ${nodes[$i-1]} "/home/pigpaxos/go/src/github.com/pigpaxos/pigpaxos/bin/start_bp.sh" "/home/ubuntu/paxi/bin/start_bp.sh"
  server_pigpaxos ${nodes[$i-1]} $i $1 $2 false 20
  echo "Started PigPaxos 1.$i"
done

