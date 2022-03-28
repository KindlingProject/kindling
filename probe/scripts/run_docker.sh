script_dir="$(dirname "$0")"
workspace_root=$(realpath "${script_dir}/../")

# Docker image information.
docker_image_with_tag="kindlingproject/kindling-compiler"

configs=(-v "$HOME/.config:/root/.config" \
  -v "$HOME/.ssh:/root/.ssh" \
  -v "$HOME/.gitconfig:/root/.gitconfig") 

IFS=' '
# Read the environment variable and set it to an array. This allows
# us to use an array access in the later command.
read -ra RUN_DOCKER_EXTRA_ARGS <<< "${RUN_DOCKER_EXTRA_ARGS}"

exec_cmd=("/usr/bin/bash")
if [ $# -ne 0 ]; then
  exec_cmd=("${exec_cmd[@]}" "-c" "$*")
fi

docker run --rm -it \
  "${configs[@]}" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "${workspace_root}/../:/kindling" \
  -w "/kindling/probe" \
  "${RUN_DOCKER_EXTRA_ARGS[@]}" \
  "${docker_image_with_tag}" \
  "${exec_cmd[@]}"
