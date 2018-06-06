load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["poseidonhttp.go"],
    importpath = "github.com/kubernetes-sigs/poseidon/pkg/poseidonhttp",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/config:go_default_library",
        "//pkg/debugutil:go_default_library",
        "//pkg/firmament:go_default_library",
        "//pkg/metrics:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus/promhttp:go_default_library",
    ],
)
