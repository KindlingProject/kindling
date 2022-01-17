script_dir="$(dirname "$0")"
run_docker_script="${script_dir}/run_docker.sh"

# This adds all the privileged flags necessary to run BPF code.
bpf_flags=(--privileged \
  -d \
  -v /:/host \
  -v /sys:/sys \
  -v /var/lib/docker:/var/lib/docker \
  "--pid=host" \
  --env "PL_HOST_PATH=/host")

RUN_DOCKER_EXTRA_ARGS="${bpf_flags[*]}"  "${run_docker_script}" "$@"
