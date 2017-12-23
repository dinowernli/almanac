load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = glob(
        ["*.go"],
        exclude = ["*_test.go"],
    ),
    visibility = ["//visibility:public"],
    deps = [
        "//proto:go_default_library",
        "@com_github_blevesearch_bleve//:go_default_library",
        "@com_github_blevesearch_bleve//document:go_default_library",
        "@com_github_blevesearch_bleve//index:go_default_library",
        "@com_github_blevesearch_bleve//index/store:go_default_library",
        "@com_github_blevesearch_bleve//mapping:go_default_library",
        "@com_github_blevesearch_bleve//registry:go_default_library",
        "@com_github_blevesearch_bleve//search/query:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_x_net//context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    size = "small",
    srcs = glob(["*_test.go"]),
    library = ":go_default_library",
    deps = [
        "@com_github_stretchr_testify//assert:go_default_library",
    ],
)
