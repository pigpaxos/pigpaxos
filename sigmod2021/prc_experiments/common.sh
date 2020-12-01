#!/usr/bin/env bash

# specify actual IP addresses
va_client=''
va_client2=''

va1=''
va2=''
va3=''
va4=''
va5=''

va6=''
va7=''
va8=''
va9=''

va10=''
va11=''
va12=''
va13=''
va14=''
va15=''
va16=''
va17=''
va18=''
va19=''
va20=''
va21=''
va22=''
va23=''
va24=''
va25=''

# used for WAN
ca1=''
ca2=''
ca3=''
ca4=''
ca5=''
ca_client=''

or1=''
or2=''
or3=''
or4=''
or5=''
or_client=''

nodes=($va1 $va2 $va3 $va4 $va5 $va6 $va7 $va8 $va9 $va10 $va11 $va12 $va13 $va14 $va15 $va16 $va17 $va18 $va19 $va20 $va21 $va22 $va23 $va24 $va25)
nodes5=($va1 $va2 $va3 $va4 $va5 )
nodes9=($va1 $va2 $va3 $va4 $va5 $va6 $va7 $va8 $va9 )
nodes11=($va1 $va2 $va3 $va4 $va5 $va6 $va7 $va8 $va9 $va10 $va11)

nodes5va=($va1 $va2 $va3 $va4 $va5)
nodes5ca=($ca1 $ca2 $ca3 $ca4 $ca5)
nodes5or=($or1 $or2 $or3 $or4 $or5)

c2="cd /home/ubuntu/paxi/bin; ./client -id=1.1 -log_level=info > logs/client.out 2>&1 &"

c2va="cd /home/ubuntu/paxi/bin; ./client -id=1.1 -log_level=info -algorithm=epaxos > logs/client.out 2>&1 &"
c2ca="cd /home/ubuntu/paxi/bin; ./client -id=2.1 -log_level=info -algorithm=epaxos > logs/client.out 2>&1 &"
c2or="cd /home/ubuntu/paxi/bin; ./client -id=3.1 -log_level=info -algorithm=epaxos > logs/client.out 2>&1 &"

server_fast_flex_paxos () {
	ssh -i va.pem ubuntu@$1 "cd /home/ubuntu/paxi/bin; ./start_ffp.sh 1.$2 $3 $4 $5 $6 &"
}

server_fast_paxos () {
	ssh -i va.pem ubuntu@$1 "cd /home/ubuntu/paxi/bin; ./start_fp.sh 1.$2 $3 &"
}

server_pigpaxos () {
  # $2 - id
  # $3 - # of relay groups
  # $4 - partial repsonse slack
  # $5 - whether to use static relays
  # $6 - default timeout
	ssh -i va.pem ubuntu@$1 "cd /home/ubuntu/paxi/bin; chmod 755 start_bp.sh; ./start_bp.sh 1.$2 $3 $4 $5 $6 &"
}

server_epaxos () {
	ssh -i va.pem ubuntu@$1 "cd /home/ubuntu/paxi/bin; chmod 755 start_ep.sh; ./start_ep.sh $2 $3 &"
}

client () {
	ssh -i va.pem ubuntu@$va_client $c2 &
}

client_wan () {
	ssh -i va.pem ubuntu@$va_client $c2va &
	ssh -i ca.pem ubuntu@$ca_client $c2ca &
	ssh -i or.pem ubuntu@$or_client $c2or &
}

upload_all () {
	ssh -i va.pem ubuntu@$1 "rm -r $3; mkdir $3"
	scp -i va.pem -r $2 ubuntu@$1:$3
}

upload_one () {
	scp -i va.pem $2 ubuntu@$1:$3
}

download () {
	scp -i va.pem ubuntu@$1:$2 $3
}