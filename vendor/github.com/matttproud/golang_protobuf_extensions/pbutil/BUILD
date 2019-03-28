load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "decode.go",
        "doc.go",
        "encode.go",
    ],
    importmap = "github.com/kubernetes-sigs/poseidon/vendor/github.com/matttproud/golang_protobuf_extensions/pbutil",
    importpath = "github.com/matttproud/golang_protobuf_extensions/pbutil",
    visibility = ["//visibility:public"],
    deps = ["@com_github_golang_protobuf//proto:go_default_library"],
)
