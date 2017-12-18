load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_test")

go_binary(
    name = "main",
    srcs = glob(
        ["*.go"],
        exclude = ["*_test.go"],
    ),
    deps = [
        "//appender:go_default_library",
        "//discovery:go_default_library",
        "//index:go_default_library",
        "//mixer:go_default_library",
        "//proto:go_default_library",
        "//storage:go_default_library",
        "@com_github_blevesearch_bleve//:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_x_net//context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    size = "small",
    srcs = glob(["*_test.go"]),
    deps = [
        "@com_github_stretchr_testify//assert:go_default_library",
    ],
)
