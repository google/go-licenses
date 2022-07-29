// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Example module that imports a module in a subdir of a repo.
package main

import (
	_ "cloud.google.com/go/storage"
	// An example for "support modules not at root":
	// ```
	// $ go-licenses csv cloud.google.com/go/storage
	// ...
	// cloud.google.com/go/storage, https://github.com/googleapis/google-cloud-go/blob/storage/v1.10.0/storage/LICENSE, Apache-2.0
	// ...
	// ```
	// Note the URL https://github.com/googleapis/google-cloud-go/blob/storage/v1.10.0/storage/LICENSE is broken, the correct URL should be https://github.com/googleapis/google-cloud-go/blob/storage/v1.10.0/LICENSE. The problem is caused by the fact that:
	// * for modules in a subdir of a repo, when go caches module files and found the submodule does not have a LICENSE file, it "magically" copies LICENSE file from root folder to the sub-module. e.g. https://github.com/googleapis/google-cloud-go/tree/storage/v1.10.0/storage
	// * therefore, go-licenses finds a LICENSE file at root of submodule and tries to guess its remote URL as root of submodule, while the actual LICENSE file is at root of repo
	// Note, adopting pkgsite/source allowed us to get the correct tag `storage/v1.10.0` for this repo, but we still hit this LICENSE file path problem.
	//
	// More context in: https://github.com/google/go-licenses/issues/73#issuecomment-1019453152
)

func main() {
}
