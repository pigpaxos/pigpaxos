## PigPaxos?

PigPaxos is a Multi-Paxos variant optimized for high number of nodes and high throughput. PigPaxos achieves this by offloading much of the communication from the dedicated leader onto the follower using the Pig communicaiton approah. Pig allows a fan-out/fan-in communication pairs to flow through a relay node. The realy disseminates the message on the fan-out path and aggragets teh responses on the fan-in side of the communication. 

## This Repo

This version of the code is anonimized for SIGMOD 2021 submission.

