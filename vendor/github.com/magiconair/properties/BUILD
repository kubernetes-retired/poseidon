load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "decode.go",
        "doc.go",
        "integrate.go",
        "lex.go",
        "load.go",
        "parser.go",
        "properties.go",
        "rangecheck.go",
    ],
    importmap = "github.com/kubernetes-sigs/poseidon/vendor/github.com/magiconair/properties",
    importpath = "github.com/magiconair/properties",
    visibility = ["//visibility:public"],
)
