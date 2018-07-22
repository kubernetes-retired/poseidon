load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["teststale.go"],
    importpath = "k8s.io/kubernetes/hack/cmd/teststale",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/golang/glog:go_default_library"],
)

go_binary(
    name = "teststale",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
