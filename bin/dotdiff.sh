#!/bin/bash

# compare the dot output files
N=$1
dir=./opera_images
FILE=DAG-EPOCH-1

diffdir=${dir}/diff
mkdir -p ${diffdir}

for i in $(seq $((N-1)))
do
	echo $i
	f=${dir}/$i/${FILE}
	fo=${dir}/$((i+1))/${FILE}	
	difffile="${diffdir}/$i-$((i+1))".diff
	
	diff ${f}.dot ${fo}.dot > ${difffile}
done

