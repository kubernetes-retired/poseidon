load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "lru.go",
        "lru_interface.go",
    ],
    importmap = "github.com/kubernetes-sigs/poseidon/vendor/github.com/hashicorp/golang-lru/simplelru",
    importpath = "github.com/hashicorp/golang-lru/simplelru",
    visibility = ["//visibility:public"],
)
