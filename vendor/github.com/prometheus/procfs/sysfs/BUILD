load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "fs.go",
        "net_class.go",
    ],
    importpath = "github.com/prometheus/procfs/sysfs",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/prometheus/procfs/bcache:go_default_library",
        "//vendor/github.com/prometheus/procfs/xfs:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "fs_test.go",
        "net_class_test.go",
    ],
    embed = [":go_default_library"],
)
