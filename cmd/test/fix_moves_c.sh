#!/bin/bash


#$ -M rnahm@nd.edu   # Email address for job notification
#$ -m abe            # Send mail when job begins, ends and aborts
#$ -pe smp 2         # Specify parallel environment and legal core size
#$ -q long           # Specify queue
#$ -N fix_c       # Specify job name


module use -a ~/privatemodules
module load golang/1.20      # Required modules

 # Application to execute
cd ~/distsys-chess-engine
./bin/test fix_moves_long 1  1000 10 1
./bin/test fix_moves_long 2  1000 10 1
./bin/test fix_moves_long 4  1000 10 1
./bin/test fix_moves_long 8  1000 10 1
./bin/test fix_moves_long 16 1000 10 1
