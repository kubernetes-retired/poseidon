load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "interrupt_handler.go",
        "sigquit_swallower_unix.go",
        "sigquit_swallower_windows.go",
    ],
    importmap = "github.com/kubernetes-sigs/poseidon/vendor/github.com/onsi/ginkgo/ginkgo/interrupthandler",
    importpath = "github.com/onsi/ginkgo/ginkgo/interrupthandler",
    visibility = ["//visibility:public"],
)
