package licenses

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/licenseclassifier/v2/assets"
)

func TestTypes(t *testing.T) {
	classifier, err := assets.DefaultClassifier()
	if err != nil {
		t.Fatal(err)
	}

	fv := reflect.ValueOf(classifier).Elem().FieldByName("docs")

	uniqueNames := make(map[string]struct{})
	for _, key := range fv.MapKeys() {
		keyValue := key.String()

		splits := strings.Split(keyValue, "/")
		_, name, _ := splits[0], splits[1], splits[2]

		if _, ok := typeMap[name]; !ok {
			t.Errorf("TypeMap does not contain %s", name)
		}

		uniqueNames[name] = struct{}{}
	}

	if len(typeMap) > len(uniqueNames) {
		t.Errorf("typeMap contains too many types: %d > %d", len(typeMap), len(uniqueNames))

		// Print out the missing types
		for name := range uniqueNames {
			delete(typeMap, name)
		}

		for name := range typeMap {
			t.Errorf("typeMap contains unknown type %s", name)
		}
	}
}

func TestTypeString(t *testing.T) {
	if got := Unknown.String(); got != "unknown" {
		t.Errorf("Unknown.String() = %q, want %q", got, "unknown")
	}
	if got := Restricted.String(); got != "restricted" {
		t.Errorf("Restricted.String() = %q, want %q", got, "restricted")
	}
	if got := Permissive.String(); got != "permissive" {
		t.Errorf("Permissive.String() = %q, want %q", got, "permissive")
	}
}

func TestLicenseType(t *testing.T) {
	if got := LicenseType("Apache-2.0"); got != Notice {
		t.Errorf("LicenseType(Apache-2.0) = %v, want %v", got, Notice)
	}
	if got := LicenseType("MIT"); got != Notice {
		t.Errorf("LicenseType(MIT) = %v, want %v", got, Notice)
	}
	if got := LicenseType("GPL-3.0"); got != Restricted {
		t.Errorf("LicenseType(GPL-3.0) = %v, want %v", got, Restricted)
	}
	if got := LicenseType("NonExistentLicense"); got != Unknown {
		t.Errorf("LicenseType(NonExistentLicense) = %v, want %v", got, Unknown)
	}
}

