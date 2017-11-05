http_archive(
    name = "io_bazel_rules_go",
    sha256 = "91fca9cf860a1476abdc185a5f675b641b60d3acf0596679a27b580af60bf19c",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.7.0/rules_go-0.7.0.tar.gz",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_repository")
load("@io_bazel_rules_go//proto:def.bzl", "proto_register_toolchains")

go_rules_dependencies()

go_register_toolchains()

proto_register_toolchains()

go_repository(
    name = "com_github_blevesearch_bleve",
    commit = "6eea5b78da004393b1d06b8c88d1bed9ca0a94b2",
    importpath = "github.com/blevesearch/bleve",
)

go_repository(
    name = "com_github_blevesearch_segment",
    commit = "db70c57796cc8c310613541dfade3dce627d09c7",
    importpath = "github.com/blevesearch/segment",
)

go_repository(
    name = "com_github_boltdb_bolt",
    commit = "144418e1475d8bf7abbdc48583500f1a20c62ea7",
    importpath = "github.com/boltdb/bolt",
)

go_repository(
    name = "com_github_steveyen_gtreap",
    commit = "0abe01ef9be25c4aedc174758ec2d917314d6d70",
    importpath = "github.com/steveyen/gtreap",
)

go_repository(
    name = "com_github_blevesearch_go_porterstemmer",
    commit = "23a2c8e5cf1f380f27722c6d2ae8896431dc7d0e",
    importpath = "github.com/blevesearch/go-porterstemmer",
)
