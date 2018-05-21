load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["poseidon.go"],
    importpath = "github.com/kubernetes-sigs/poseidon/cmd/poseidon",
    visibility = ["//visibility:private"],
    deps = [
        "//pkg/config:go_default_library",
        "//pkg/debugutil:go_default_library",
        "//pkg/firmament:go_default_library",
        "//pkg/k8sclient:go_default_library",
        "//pkg/stats:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/wait:go_default_library",
    ],
)

go_binary(
    name = "poseidon",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
