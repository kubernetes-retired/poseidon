load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "poseidonstats.pb.go",
        "poseidonstats_service_mock.go",
        "stats.go",
    ],
    importpath = "github.com/kubernetes-sigs/poseidon/pkg/stats",
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/firmament:go_default_library",
        "//pkg/k8sclient:go_default_library",
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/golang/mock/gomock:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["stats_test.go"],
    embed = [":go_default_library"],
    deps = [
        "//pkg/firmament:go_default_library",
        "//vendor/github.com/golang/mock/gomock:go_default_library",
    ],
)
