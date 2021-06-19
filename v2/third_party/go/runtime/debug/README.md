# go/runtime/debug

Vendored from <https://cs.opensource.google/go/go/+/refs/tags/go1.16.6:src/runtime/debug/mod.go>.

Local modifications:

* The original debug.ReadBuildInfo function reads build info for the currently
  running go binary. The new debug.ParseBuildInfo function is modified from it to
  parse build info from `go version -m <binary_path>` command output.
  There are some minor format differences.
