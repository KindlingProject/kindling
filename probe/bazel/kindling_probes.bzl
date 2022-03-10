load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kindling_probes():
    http_file(
        name = "kindling_falcolib_probe_tar_gz",
        urls = ["https://k8s-bpf-probes-public.oss-cn-hangzhou.aliyuncs.com/kindling-falcolib-probe.tar.gz"],
        sha256 = "fc6b44cfa36ad7cb1c532a07c0e5a761b289d696a33299af7293878003db74ac",
        downloaded_file_path = "kindling-falcolib-probe.tar.gz",
    )
