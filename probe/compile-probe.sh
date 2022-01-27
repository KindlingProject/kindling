#!/bin/bash

# file structure:
#		- driver
# 			- Makefile (kernel modules)
#			- *.h, *.c
#			- bpf
#				- *.h, *.c
#				- Makefile (bpf probe)
#		- compile-probe.sh: script for compiling
#   - probe: contains all probes after compiling
PROBE_NAME=hcmine-probe
DST=kindling-probe
DIR=`pwd`
if [ ! -d $DST  ];then
  mkdir $DST
fi
for version in `ls /lib/modules/`
do
	cd $DIR/src/driver
	echo Compile probe for $version
	src=/lib/modules/$version/build
	# compile kernel modules
	make KERNELDIR=$src
	mv $PROBE_NAME.ko $DIR/$DST/$version.ko
	make KERNELDIR=$src clean
	# compile bpf modules
	cd bpf
	make KERNELDIR=$src
	mv probe.o $DIR/$DST/$version.o
	make KERNELDIR=$src clean
done