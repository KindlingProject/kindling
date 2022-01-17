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

DIR=`pwd`
for version in `ls /lib/modules/`
do
	cd $DIR/driver
	echo Compile probe for $version
	src=/lib/modules/$version/build
	# compile kernel modules
	make KERNELDIR=$src
	mv hcmine-probe.ko $DIR/deploy/probe/$version.ko
	make KERNELDIR=$src clean
	# compile bpf modules
	cd bpf
	make KERNELDIR=$src
	mv probe.o $DIR/deploy/probe/$version.o
	make KERNELDIR=$src clean
done