GOOGLEAPIS_GOOGLE_API_BUILD_CONTENTS = """
load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")

package(default_visibility = ["//visibility:public"])

proto_library(
    name = "api_proto",
    srcs = [
        "annotations.proto",
        "http.proto",
    ],
    deps = ["@com_google_protobuf//:descriptor_proto"],
)

go_proto_library(
    name = "api_go_proto",
    importpath = "google.golang.org/genproto/googleapis/api/annotations",
    proto = ":api_proto",
    deps = ["@com_github_golang_protobuf//protoc-gen-go/descriptor:go_default_library"],
)

go_library(
    name = "go_default_library",
    embed = [":api_go_proto"],
    importpath = "google.golang.org/genproto/googleapis/api/annotations",
)
"""

def _googleapis_repository_impl(ctx):
    googleapis_commit = "e1c0c726290a55065c0c46a62dacc9372939973b"
    ctx.download_and_extract(
        url = "https://github.com/googleapis/googleapis/archive/{commit}.tar.gz".format(
            commit = googleapis_commit,
        ),
        sha256 = "9508971cb4a7c0fe03bc1bfafbd0abc9654c80b4c70e360a6c534938d06d8fb9",
        stripPrefix = "googleapis-{}".format(googleapis_commit),
    )

    ctx.file("google/api/BUILD.bazel", GOOGLEAPIS_GOOGLE_API_BUILD_CONTENTS)


_googleapis_repository = repository_rule(
    implementation = _googleapis_repository_impl,
)


def repositories():
    _googleapis_repository(name = "com_github_googleapis_googleapis")
