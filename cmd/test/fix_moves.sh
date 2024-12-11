#!/bin/bash


#$ -pe smp 2         # Specify parallel environment and legal core size
#$ -q long           # Specify queue
#$ -N short-buf       # Specify job name
#$ -t 1-16	 	# Specify number of copies

module use -a ~/privatemodules
module load golang/1.20      # Required modules

 # Application to execute
cd ~/distsys-chess-engine
formatted_id=$(printf "%02d" "$((SGE_TASK_ID - 1))")
./bin/server "short-buf-$formatted_id"
