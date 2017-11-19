load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "main",
    srcs = glob(["*.go"]),
    deps = [
        "//index:go_default_library",
        "//proto:go_default_library",
        "//storage:go_default_library",
        "@com_github_blevesearch_bleve//:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_x_net//context:go_default_library",
    ],
)
