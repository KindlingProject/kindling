#!/bin/sh

/pl/kindling-probe-loader

if [ -f "/opt/probe.o" ]; then
	export SYSDIG_BPF_PROBE="/opt/probe.o"
fi

exec /pl/kindling_probe "$@"
