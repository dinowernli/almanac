http_archive(
    name = "org_pubref_rules_protobuf",
    strip_prefix = "rules_protobuf-0.8.1",
    urls = ["https://github.com/pubref/rules_protobuf/archive/v0.8.1.zip"],
)

load("@org_pubref_rules_protobuf//java:rules.bzl", "java_proto_repositories")

java_proto_repositories()

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "d322432e804dfa7936c1f38c24b00f1fb71a0be090a1273d7100a1b4ce281ee7",
    strip_prefix = "rules_go-a390e7f7eac912f6e67dc54acf67aa974d05f9c3",
    urls = [
        "http://bazel-mirror.storage.googleapis.com/github.com/bazelbuild/rules_go/archive/a390e7f7eac912f6e67dc54acf67aa974d05f9c3.tar.gz",
        "https://github.com/bazelbuild/rules_go/archive/a390e7f7eac912f6e67dc54acf67aa974d05f9c3.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:def.bzl", "go_repositories", "new_go_repository")

go_repositories()

