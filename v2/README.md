# go-licenses v2

A tool to automate license management workflow for go module project's dependencies and transitive dependencies.

## **THIS IS STILL UNDER DEVELOPMENT**

The v2 package is being developed and currently incomplete, @Bobgy is
upstreaming changes from his fork in <https://github.com/Bobgy/go-licenses/blob/main/v2>.

Tracking issue where you can find the roadmap and progress:
<https://github.com/google/go-licenses/issues/70>.

The major changes from v1 are:

* V2 only supports go modules, it can get license URL for modules without a need for you to vendor your dependencies.
* V2 does not assume each module has a single license, v2 will scan all the files for each module to find licenses.
