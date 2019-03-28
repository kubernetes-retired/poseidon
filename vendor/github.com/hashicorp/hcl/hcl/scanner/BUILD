load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["scanner.go"],
    importmap = "github.com/kubernetes-sigs/poseidon/vendor/github.com/hashicorp/hcl/hcl/scanner",
    importpath = "github.com/hashicorp/hcl/hcl/scanner",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/hashicorp/hcl/hcl/token:go_default_library"],
)
