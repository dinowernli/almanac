load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_binary(
    name = "main",
    srcs = [
        "cluster.go",
        "main.go",
    ],
    deps = [
        "//pkg/http:go_default_library",
        "//pkg/index:go_default_library",
        "//pkg/service/appender:go_default_library",
        "//pkg/service/discovery:go_default_library",
        "//pkg/service/ingester:go_default_library",
        "//pkg/service/mixer:go_default_library",
        "//pkg/storage:go_default_library",
        "//proto:go_default_library",
        "@com_github_alecthomas_kingpin//:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_sirupsen_logrus//:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//codes:go_default_library",
        "@org_golang_x_net//context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    size = "small",
    srcs = glob([
        "*_test.go",
        "cluster.go",
    ]),
    deps = [
        "//pkg/index:go_default_library",
        "//pkg/service/appender:go_default_library",
        "//pkg/service/discovery:go_default_library",
        "//pkg/service/ingester:go_default_library",
        "//pkg/service/mixer:go_default_library",
        "//pkg/storage:go_default_library",
        "//proto:go_default_library",
        "@com_github_blevesearch_bleve//:go_default_library",
        "@com_github_sirupsen_logrus//:go_default_library",
        "@com_github_stretchr_testify//assert:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_x_net//context:go_default_library",
    ],
)
