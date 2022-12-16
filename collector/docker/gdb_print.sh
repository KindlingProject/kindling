#!/bin/sh

binname=$1
corefile=$2

(gdb $binname $corefile > ./$corefile.log 2>&1) <<GDBEOF
thread apply all bt
GDBEOF

cat ./$corefile.log
rm -f ./$corefile.log ./$corefile