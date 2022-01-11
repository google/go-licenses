# go/pkgsite/source

Vendored from <https://go.googlesource.com/pkgsite/+/ef8047b111963f61f5a0b3ae5b464cdc18dc5f74/internal/source>,
because the source package is internal, and there's no plan to move it out anytime soon: <https://github.com/golang/go/issues/40477#issuecomment-868532845>.

Local modifications:

- Update import paths.
- Removed unused functions from pkgsite/internal/stdlib, pkgsite/internal/derrors,
  pkgsite/internal/version to avoid other dependencies.
- For pkgsite/internal/source, switched to use go log package, because glog conflicts with a test
  dependency that also defines the "v" flag.
