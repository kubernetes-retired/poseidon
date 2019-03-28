load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "decode_hooks.go",
        "error.go",
        "mapstructure.go",
    ],
    importmap = "github.com/kubernetes-sigs/poseidon/vendor/github.com/mitchellh/mapstructure",
    importpath = "github.com/mitchellh/mapstructure",
    visibility = ["//visibility:public"],
)
