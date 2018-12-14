load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:exclude third_party

gazelle(
    name = "gazelle_diff",
    mode = "diff",
    prefix = "github.com/grpc-ecosystem/grpc-gateway",
)

gazelle(
    name = "gazelle_fix",
    mode = "fix",
    prefix = "github.com/grpc-ecosystem/grpc-gateway",
)

package_group(
    name = "generators",
    packages = [
        "//protoc-gen-grpc-gateway/...",
        "//protoc-gen-swagger/...",
    ],
)
