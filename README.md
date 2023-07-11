# bouncer

![app-pipeline](https://github.com/sulaiman-coder/gobouncer/workflows/app-pipeline/badge.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/sulaiman-coder/gobouncer)

A go dependency license checker.

*This is thin a wrapper around [google's license classifier](https://www.github.com/google/licenseclassifier) forked from [go-license](https://www.github.com/google/gobouncer) with a few extra options.*

## Installation

```bash
# install the latest version to ./bin
curl -sSfL https://raw.githubusercontent.com/sulaiman-coder/gobouncer/master/bouncer.sh | sh 

# install a specific version to another directory
curl -sSfL https://raw.githubusercontent.com/sulaiman-coder/gobouncer/master/bouncer.sh | sh -s -- -b ./path/to/bin v1.26.0
```

## Usage

```bash
# list the licenses of all of your dependencies...
bouncer list                        # ... from ./go.mod
bouncer list ~/some/path            # ... from ~/some/path/go.mod
bouncer list github.com/some/repo   # ... from a remote repo

# pass/fail of user-specified license restrictions (by .bouncer.yaml)
bouncer check
bouncer check ~/some/path
bouncer check github.com/some/repo
```

The `.bouncer.yaml` can specify a simple allow-list or deny-list license name regex patterns (by SPDX name):

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
