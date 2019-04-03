load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "affinity.pb.go",
        "avoid_pods_annotation.pb.go",
        "coco_interference_scores.pb.go",
        "firmament_client.go",
        "firmament_scheduler.pb.go",
        "firmament_scheduler_mock.go",
        "job_desc.pb.go",
        "label.pb.go",
        "label_selector.pb.go",
        "node_affinity.pb.go",
        "pod_affinity.pb.go",
        "pod_anti_affinity.pb.go",
        "reference_desc.pb.go",
        "resource_desc.pb.go",
        "resource_stats.pb.go",
        "resource_topology_node_desc.pb.go",
        "resource_vector.pb.go",
        "scheduling_delta.pb.go",
        "taints.pb.go",
        "task_desc.pb.go",
        "task_final_report.pb.go",
        "task_stats.pb.go",
        "tolerations.pb.go",
        "whare_map_stats.pb.go",
    ],
    importpath = "github.com/kubernetes-sigs/poseidon/pkg/firmament",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/golang/mock/gomock:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc/grpclog:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["firmament_client_test.go"],
    embed = [":go_default_library"],
    deps = ["//vendor/github.com/golang/mock/gomock:go_default_library"],
)
