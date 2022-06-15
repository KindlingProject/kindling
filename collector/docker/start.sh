
/usr/bin/kindling-probe-loader

if [ -f "/opt/probe.o" ]; then
	export SYSDIG_BPF_PROBE="/opt/probe.o"
fi

/usr/bin/kindling-collector --config=/etc/kindling/config/kindling-collector-config.yml