load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = glob(
        ["*.go"],
        exclude = ["*_test.go"],
    ),
    visibility = ["//visibility:public"],
    deps = [
        "//http:go_default_library",
        "//index:go_default_library",
        "//proto:go_default_library",
        "//service/discovery:go_default_library",
        "//storage:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_sirupsen_logrus//:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//codes:go_default_library",
        "@org_golang_x_net//context:go_default_library",
    ],
)
