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

// This is an example of a go module that vendors its dependencies using `go mod vendor`.
package main

import (
	_ "github.com/Azure/azure-sdk-for-go"
	_ "github.com/Azure/go-autorest/autorest"
	_ "github.com/Azure/go-autorest/autorest/adal"
	_ "github.com/Azure/go-autorest/autorest/to"
	_ "github.com/akamai/AkamaiOPEN-edgegrid-golang"
	_ "github.com/cpu/goacmedns"
	_ "github.com/digitalocean/godo"
	_ "github.com/go-ldap/ldap/v3"
	_ "github.com/go-logr/logr"
	_ "github.com/google/gofuzz"
	_ "github.com/hashicorp/vault/api"
	_ "github.com/hashicorp/vault/sdk/logical"
	_ "github.com/kr/pretty"
	_ "github.com/miekg/dns"
	_ "github.com/onsi/ginkgo/v2"
	_ "github.com/onsi/gomega"
	_ "github.com/pavlo-v-chernykh/keystore-go/v4"
	_ "github.com/pkg/errors"
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/spf13/cobra"
	_ "github.com/spf13/pflag"
	_ "github.com/stretchr/testify"
	_ "golang.org/x/crypto/sha3"
	_ "golang.org/x/exp/slices"
	_ "golang.org/x/oauth2"
	_ "golang.org/x/sync/errgroup"
	_ "gomodules.xyz/jsonpatch/v2"
	_ "google.golang.org/api"
	_ "k8s.io/api"
	_ "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	_ "k8s.io/apimachinery"
	_ "k8s.io/apiserver"
	_ "k8s.io/client-go"
	_ "k8s.io/code-generator"
	_ "k8s.io/component-base"
	_ "k8s.io/klog/v2"
	_ "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	_ "k8s.io/utils/pointer"
	_ "sigs.k8s.io/controller-runtime"
	_ "sigs.k8s.io/gateway-api/apis/v1beta1"
	_ "sigs.k8s.io/structured-merge-diff/v4/fieldpath"
)

func main() {
}
