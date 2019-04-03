package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = ["hash.go"],
    importmap = "github.com/kubernetes-sigs/poseidon/vendor/k8s.io/kubernetes/pkg/util/hash",
    importpath = "k8s.io/kubernetes/pkg/util/hash",
    deps = ["//vendor/github.com/davecgh/go-spew/spew:go_default_library"],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)
