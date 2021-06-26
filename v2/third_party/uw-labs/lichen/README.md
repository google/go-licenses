# uw-labs/lichen

uw-labs/lichen is a license tool to analyze go binaries.
We vendor one internal package of uw-labs/lichen which can parse modules from
outputs of `go version -m` command.

Refer to [go_binary.go](../../../gocli/go_binary.go) for why we chose to vendor
this package.

## Metadata

Upstream repo URL: https://github.com/uw-labs/lichen.
The vendored commit is be9752894a5958f6ba7be9e05dc370b7a73b58db.
Last upgrade date: 2021-4-11

## Local modifications

* Changed import path to github.com/google/go-licenses/v2/third_party/uw-labs/lichen.
* Removed unused code in google/go-licenses.
