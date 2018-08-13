load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "context.go",
        "doc.go",
        "http.go",
        "logger.go",
        "trace.go",
        "util.go",
        "version.go",
    ],
    importpath = "github.com/docker/distribution/context",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/Sirupsen/logrus:go_default_library",
        "//vendor/github.com/docker/distribution/uuid:go_default_library",
        "//vendor/github.com/gorilla/mux:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
