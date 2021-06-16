// Copyright 2021 Google LLC
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

package deps

type GoModule struct {
	// go import path, example: github.com/google/licenseclassifier/v2
	ImportPath string
	// version, example: v1.2.3, v0.0.0-20201021035429-f5854403a974
	Version string
	// local directory of dependency's source code, example on MacOS:
	// /Users/username/go/pkg/mod/github.com/!puerkito!bio/goquery@v1.6.1
	SrcDir string
}
