load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = glob(["*.go"]),
    visibility = ["//visibility:public"],
    deps = [
        "//index:go_default_library",
        "//proto:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
    ],
)
