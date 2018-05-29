load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["metrics.go"],
    importpath = "github.com/kubernetes-sigs/poseidon/pkg/metrics",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/prometheus/client_golang/prometheus:go_default_library"],
)
