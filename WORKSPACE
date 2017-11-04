http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.7.0/rules_go-0.7.0.tar.gz",
    sha256 = "91fca9cf860a1476abdc185a5f675b641b60d3acf0596679a27b580af60bf19c",
)
load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "new_go_repository")
go_rules_dependencies()
go_register_toolchains()

new_go_repository(
    name = "com_github_blevesearch_bleve",
    importpath = "github.com/blevesearch/bleve",
    tag = "v0.5.0",
)

