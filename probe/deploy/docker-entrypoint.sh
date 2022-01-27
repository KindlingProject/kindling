#!/bin/sh

/pl/kindling-probe-loader

if [ -f "/opt/probe.o" ]; then
	export HCMINE_BPF_PROBE="/opt/probe.o"
fi

exec /pl/kindling_probe "$@"
