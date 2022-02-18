load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kindling_probes():
    http_file(
        name = "kindling_falcolib_probe_tar_gz",
        urls = ["https://k8s-bpf-probes-public.oss-cn-hangzhou.aliyuncs.com/kindling-falcolib-probe.tar.gz"],
        sha256 = "d1962dddb86ef44d89b490b7ca37c07c283feeba73a235e9cdba42c556ffed6b",
        downloaded_file_path = "kindling-falcolib-probe.tar.gz",
    )
