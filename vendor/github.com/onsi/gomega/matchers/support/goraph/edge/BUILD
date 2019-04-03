load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["edge.go"],
    importmap = "github.com/kubernetes-sigs/poseidon/vendor/github.com/onsi/gomega/matchers/support/goraph/edge",
    importpath = "github.com/onsi/gomega/matchers/support/goraph/edge",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/onsi/gomega/matchers/support/goraph/node:go_default_library"],
)
