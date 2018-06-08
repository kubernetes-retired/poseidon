load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "bcache.go",
        "get.go",
    ],
    importpath = "github.com/prometheus/procfs/bcache",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["get_test.go"],
    embed = [":go_default_library"],
)
