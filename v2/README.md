# go-licenses

## **THIS IS STILL UNDER DEVELOPMENT**

A tool to automate license management workflow for go module project's dependencies and transitive dependencies.

## Install

Download the released package and install it to your PATH:
TODO: udpate URL after release.

```bash
curl -LO download-url/go-licenses-linux.tar.gz
tar xvf go-licenses-linux.tar.gz
sudo mv go-licenses/* /usr/local/bin/
# or move the content to anywhere in PATH
```

## Output Example

<!-- TODO: update NOTICES folder of this repo. -->
<!-- [NOTICES folder](./NOTICES) is an example of generated NOTICES for go-licenses tool itself. -->

Examples used in Kubeflow Pipelines:

* [go-licenses.yaml (config file)](https://github.com/kubeflow/pipelines/blob/master/v2/go-licenses.yaml)
* [license_info.csv (generated)](https://github.com/kubeflow/pipelines/blob/master/v2/third_party/license_info.csv)
* [NOTICES/licenses.txt (generated)](https://github.com/kubeflow/pipelines/blob/master/v2/third_party/NOTICES/licenses.txt)

## Usage

### One-off License Update

1. Get version of the repo you need licenses info:

    ```bash
    git clone <go-mod-repo-you-need-license-info>
    cd <go-mod-repo-you-need-license-info>
    git checkout <version>
    ```

1. Write down a minimal config file specifying your module name and which binary to analyze:

    ```yaml
    module:
        go:
            module: github.com/google/go-licenses/v2
            path: .
            binary:
                path: dist/linux/go-licenses
    ```

1. Get dependencies from go modules and generate a `license_info.csv` file of their licenses:

    ```bash
    go-licenses csv
    ```

    The csv file has three columns: `depdency`, `license download url` and inferred `license type`.

    Note, the format is consistent with [google/go-licenses](https://github.com/google/go-licenses).

1. The tool may fail to identify:

    * Download url of a license: they will be left out in the csv.
    * SPDX ID of a license: they will be named `Unknown` in the csv.

    Please check them manually and update your `go-licenses.yaml` config to fix them, refer to [the example](./go-licenses.yaml). After your config fix, re-run the tool to generate lists again:

    ```bash
    go-licenses csv
    ```

    Iterate until you resolved all license issues.

1. Download notices, licenses and source folders that should be distributed along with the built binary:

    ```bash
    go-licenses save
    ```

    Notices and licenses will be concatenated to a single file called `NOTICES/license.txt`.
    Source code folders will be copied to `NOTICES/<module/import/path>`.

    Notices folder location can be configured in [the go-licenses.yaml example](./go-licenses.yaml).

    Some licenses will be rejected based on its [license type](https://github.com/google/licenseclassifier/blob/df6aa8a2788bdf5ac382148c2453a407a29819b8/license_type.go#L341).

### Integrating in CI

Typically, I think we should check `licenses_info.csv` into source control and
download license contents when releasing.

An early idea for CI is to run a simple script:

1. clones the repo, run `go-licenses csv`.
1. verifies if generated `licenses_info.csv` if up-to-date as the version in the repo.

We might worry about flakiness, because various dependencies could be down
temporarily. Another simpler idea is to let the script do:

1. If `go.mod` has been updated, but not the license files.
1. Fails and says you should update the license files.

## Implementation Details

Rough idea of steps in the two commands.

`go-licenses csv` does the following to generate the `license_info.csv`:

1. Load `go-licenses.yaml` config file, the config file can contain
    * module name
    * built binary local path
    * module license overrides (path excludes or directly assign result license)
1. All dependencies and transitive dependencies are listed by `go version -m <binary-path>`. When a binary is built with go modules, used module info are logged inside the binary. Then we parse go CLI result to get the full list.
1. Scan licenses and report problems:
    1. Use <github.com/google/licenseclassifier/v2> detect licenses from all files of dependencies.
    1. Report an error if no license found for a dependency etc.
1. Get license public URLs:
    1. Get a dependency's github repo by fetching meta info like `curl 'https://k8s.io/client-go?go-get=1'`.
    1. Get dependency's version info from go modules metadata.
    1. Combine github repo, version and license file path to a public github URL to the license file.
1. Generate CSV output with module name, license URL and license type.
1. Report dependencies the tool failed to deal with during the process.

`go-licenses save` does the following:

1. Read from `license_info.csv` generated in `go-licenses csv`.
1. Call [github.com/google/licenseclassifier](https://github.com/google/licenseclassifier) to get license type.
1. Three types of reactions to license type:
    * Download its notice and license for all types.
    * Copy source folder for types that require redistribution of source code.
    * Reject according to <https://github.com/google/licenseclassifier/blob/df6aa8a2788bdf5ac382148c2453a407a29819b8/license_type.go#L341>.

## Credits

go-licenses/v2 is greatly inspired by

* [github.com/google/go-licenses](https://github.com/google/go-licenses) for the commands and compliance workflow
* [github.com/mitchellh/golicense](https://github.com/mitchellh/golicense) for getting modules from binary
* [github.com/uw-labs/lichen](https://github.com/uw-labs/lichen) for the vendored code to extract structured data from `go version -m` result.

## Comparison with similar tools

<!-- TODO(Bobgy): update this to a table -->

* go-licenses/v2 was greatly inspired by [github.com/google/go-licenses](https://github.com/google/go-licenses), with the differences:
  * go-licenses/v2 works better with go modules.
    * no need to vendor dependencies.
    * discovers versioned license URLs.
  * go-licenses/v2 scans all dependency files to find multiple licenses if any, while go-licenses detects by file name heuristics in local source folders and only finds one license per dependency.
  * go-licenses/v2 supports using a manually maintained config file `go-licenses.yaml`, so that we can reuse periodic license changes with existing information.
* go-licenses/v2 was mostly written before I learned [github.com/github/licensed](https://github.com/github/licensed) is a thing.
  * Similar to google/go-licenses, github/licensed only use heuristics to find licenses and assumes one license per repo.
  * github/licensed uses a different library for detecting and classifying licenses.
* go-licenses/v2 is a rewrite of [kubeflow/testing/go-license-tools](https://github.com/kubeflow/testing/tree/master/py/kubeflow/testing/go-license-tools) in go, with many improvements:
  * better & more robust github repo resolution ratio
  * better license classification rate using google/licenseclassifier/v2 (it especially handles BSD-2-Clause and BSD-3-Clause significantly better than GitHub license API).
  * automates licenses that require distributing source code with it (copied from local module src cache)
  * simpler process e2e (instead of too many intermediate steps and config files)
  * rewritten in go, so it's easier to redistribute the binary than python

## Roadmap

General directions to improve this tool:

* Build backward compatible behavior compared to google/go-licenses v1.
* Ask for more usage & feedback and improve robustness of the tool.

## TODOs

### Features

#### P0

* [ ] Use cobra to support providing the same information via argument or config.
* [ ] Implement "check" command.
* [ ] Support use-case of one modules folder with multiple binaries.
* [x] Support replace directives.
* [x] Support modules with +incompatible in their versions, ref: <https://golang.org/ref/mod#incompatible-versions>.

#### P1

* [ ] Support installation using go get.
* [ ] Refactor & improve test coverage.

#### P2

* [ ] Support auto inclusion of licenses in headers by recording start line and end line of a license detection.
* [ ] Check header licenses match their root license.
* [ ] Find better default locations of generated files.
* [ ] Improve logging format & consistency.
* [ ] Tutorial for integration in CI/CD.

## License Workflow Design Overview

This section introduces full workflow to comply with open source licenses.
In each workflow stage, we list several options and what this tool prefers.

1. List dependencies - Options
    * (Preferred) List dependencies in a go binary
    * List all go module dependencies

1. Detect licenses for a dependency
    * Files to consider - options:
        * (Preferred) Scan every file
        * Only look into common license file names like LICENSE, LICENSE.txt, COPYING, etc.
    * License classifier - options:
        * (Preferred) [google/licenseclassifier/v2](https://github.com/google/licenseclassifier/tree/main/v2)
        * [licensee](https://github.com/licensee/licensee)
        * GitHub license API
        * many other options
    * Manual configs to overcome what we cannot automate
        * (not supported yet) allowlist for licenses
        * (supported) override manually examined licenses
        * (supported) exclude self-owned proprietary dependencies
        * (supported) pin config to dependency version to avoid stale configs

1. Comply with license requirements by redistributing:
    * attribution/copyright notice
    * licenses in full text
    * dependency source code for licenses that require so
