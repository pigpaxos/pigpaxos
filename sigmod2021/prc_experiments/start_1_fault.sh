#!/usr/bin/env bash

source common.sh

# redefine nodes list to have one fewer node. It will not start and will act as a failed one. $va16 is missing
nodes=($va1 $va2 $va3 $va4 $va5 $va6 $va7 $va8 $va9 $va10 $va11 $va12 $va13 $va14 $va15 '' $va17 $va18 $va19 $va20 $va21 $va22 $va23 $va24 $va25)

arraylength=${#nodes[@]}
for (( i=1; i<${arraylength}+1; i++ ));
do
  ssh -i va.pem ubuntu@${nodes[$i-1]} 'cd /home/ubuntu/paxi/bin/; rm logs/*'
  #upload_one ${nodes[$i-1]} "/home/pigpaxos/go/src/github.com/pigpaxos/pigpaxos/bin/start_bp.sh" "/home/ubuntu/paxi/bin/start_bp.sh"
  server_pigpaxos ${nodes[$i-1]} $i $1 $2 false 20
  echo "Started PigPaxos 1.$i"
done

