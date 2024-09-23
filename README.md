# golicenses

![app-pipeline](https://github.com/khulnasoft/go-licenses/workflows/app-pipeline/badge.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/khulnasoft/go-licenses)

A go dependency license checker.

*This is thin a wrapper around [google's license classifier](https://www.github.com/google/licenseclassifier) forked from [go-license](https://www.github.com/google/go-licenses) with a few extra options.*

## Installation

```bash
# install the latest version to ./bin
curl -sSfL https://raw.githubusercontent.com/khulnasoft/go-licenses/khulnasoft/golicenses.sh | sh 

# install a specific version to another directory
curl -sSfL https://raw.githubusercontent.com/khulnasoft/go-licenses/khulnasoft/golicenses.sh | sh -s -- -b ./path/to/bin v1.26.0
```

## Usage

```bash
# list the licenses of all of your dependencies...
golicenses list                        # ... from ./go.mod
golicenses list ~/some/path            # ... from ~/some/path/go.mod
golicenses list github.com/some/repo   # ... from a remote repo

# pass/fail of user-specified license restrictions (by .golicenses.yaml)
golicenses check
golicenses check ~/some/path
golicenses check github.com/some/repo
```

The `.golicenses.yaml` can specify a simple allow-list or deny-list license name regex patterns (by SPDX name):

```bash
permit:
  - BSD.*
  - MIT.*
  - Apache.*
  - MPL.*
```

```bash
forbid:
  - GPL.*
```

```bash
ignore-packages:
  - github.com/some/repo
forbid:
  - GPL.*
```

Note: either allow or deny lists can be specified, not both.
