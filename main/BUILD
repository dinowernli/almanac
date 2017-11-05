load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "main",
    srcs = glob(["*.go"]),
    deps = [
        "//proto:go_default_library",
        "@com_github_blevesearch_bleve//:go_default_library",
        "@com_github_blevesearch_bleve//document:go_default_library",
        "@com_github_blevesearch_bleve//index:go_default_library",
        "@com_github_blevesearch_bleve//index/store:go_default_library",
        "@com_github_blevesearch_bleve//mapping:go_default_library",
        "@com_github_blevesearch_bleve//search/query:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_x_net//context:go_default_library",
    ],
)
