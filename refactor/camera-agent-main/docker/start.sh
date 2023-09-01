
/usr/bin/kindling-probe-loader

if [ -f "/opt/probe.o" ]; then
	export SYSDIG_BPF_PROBE="/opt/probe.o"
fi

/app/camera-agent --config=/app/config/camera-agent-config.yml