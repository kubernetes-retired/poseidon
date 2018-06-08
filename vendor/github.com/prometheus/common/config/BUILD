load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "config.go",
        "http_config.go",
    ],
    importpath = "github.com/prometheus/common/config",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/mwitkow/go-conntrack:go_default_library",
        "//vendor/gopkg.in/yaml.v2:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "http_config_test.go",
        "tls_config_test.go",
    ],
    data = glob(["testdata/**"]),
    embed = [":go_default_library"],
    deps = ["//vendor/gopkg.in/yaml.v2:go_default_library"],
)
