package openapi_generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUniqueItemsMigrationCompatibility(t *testing.T) {
	legacy := extensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "examples.test.local", Annotations: map[string]string{"nexus": `{}`}},
		Spec: extensionsv1.CustomResourceDefinitionSpec{
			Versions: []extensionsv1.CustomResourceDefinitionVersion{{Name: "v1", Schema: &extensionsv1.CustomResourceValidation{
				OpenAPIV3Schema: &extensionsv1.JSONSchemaProps{Type: "object", Properties: map[string]extensionsv1.JSONSchemaProps{
					"values": {Type: "array", UniqueItems: true, Items: &extensionsv1.JSONSchemaPropsOrArray{Schema: &extensionsv1.JSONSchemaProps{Type: "string"}}},
				}},
			}}},
		},
	}
	oldDir, newDir := t.TempDir(), t.TempDir()
	updated := legacy.DeepCopy()
	normalizeUniqueItems(updated.Spec.Versions[0].Schema.OpenAPIV3Schema)
	property := updated.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["values"]
	if property.UniqueItems || property.XListType == nil || *property.XListType != "set" {
		t.Fatalf("uniqueness was not preserved as a structural set: %+v", property)
	}
	for dir, crd := range map[string]*extensionsv1.CustomResourceDefinition{oldDir: &legacy, newDir: updated} {
		data, err := yaml.Marshal(crd)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "example.yaml"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := CheckBackwardCompatibility(oldDir, newDir, false); err != nil {
		t.Fatalf("equivalent uniqueness migration rejected: %v", err)
	}
}
