load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

def kindling_probes():
    http_file(
        name = "kindling_probes_tar_gz",
        urls = ["https://arms-usw-public-public.oss-us-west-1.aliyuncs.com/kindling-probe.tar.gz"],
        sha256 = "6b56304f4b7305a8778495764d97b966a778467d041e6338b6ed0f3608cd6ee2",
        downloaded_file_path = "kindling-probes.tar.gz",
    )
