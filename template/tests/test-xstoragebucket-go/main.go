// Package main generates a CompositionTest
package main

import (
	"encoding/json"
	"fmt"
	"os"

	metav1 "dev.upbound.io/models/io/k8s/meta/v1"
	metav1alpha1 "dev.upbound.io/models/io/upbound/dev/meta/v1alpha1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

func main() {
	// Define assertions for the composed resources
	// These assertions validate the correct structure and values of the resources
	// created by the compose-bucket function
	assertResources := resourcesToItems[metav1alpha1.CompositionTestSpecAssertResourcesItem](
		// Assert ResourceGroup with correct location
		map[string]interface{}{
			"apiVersion": "azure.upbound.io/v1beta1",
			"kind":       "ResourceGroup",
			"spec": map[string]interface{}{
				"forProvider": map[string]interface{}{
					"location": "eastus",
				},
			},
		},
		// Assert Storage Account with correct configuration
		map[string]interface{}{
			"apiVersion": "storage.azure.upbound.io/v1beta1",
			"kind":       "Account",
			"metadata": map[string]interface{}{
				"name": "example", // Name transformation (hyphens removed)
			},
			"spec": map[string]interface{}{
				"forProvider": map[string]interface{}{
					"accountTier":                     "Standard",
					"accountReplicationType":          "LRS",
					"location":                        "eastus",
					"infrastructureEncryptionEnabled": true,
					"blobProperties": []interface{}{
						map[string]interface{}{
							"versioningEnabled": true,
						},
					},
					"resourceGroupNameSelector": map[string]interface{}{
						"matchControllerRef": true,
					},
				},
			},
		},
		// Assert Storage Container with correct ACL mapping
		map[string]interface{}{
			"apiVersion": "storage.azure.upbound.io/v1beta1",
			"kind":       "Container",
			"spec": map[string]interface{}{
				"forProvider": map[string]interface{}{
					"containerAccessType": "blob", // ACL "public" maps to "blob"
					"storageAccountNameSelector": map[string]interface{}{
						"matchControllerRef": true,
					},
				},
			},
		},
	)
	test := metav1alpha1.CompositionTest{
		APIVersion: ptr.To(metav1alpha1.CompositionTestAPIVersionmetaDevUpboundIoV1Alpha1),
		Kind:       ptr.To(metav1alpha1.CompositionTestKindCompositionTest),
		Metadata: &metav1.ObjectMeta{
			Name: ptr.To(""),
		},
		Spec: &metav1alpha1.CompositionTestSpec{
			AssertResources: &assertResources,
			CompositionPath: ptr.To("apis/xstoragebuckets/composition.yaml"),
			XrPath:          ptr.To("examples/xstoragebuckets/example.yaml"),
			XrdPath:         ptr.To("apis/xstoragebuckets/definition.yaml"),
			TimeoutSeconds:  ptr.To(120),
			Validate:        ptr.To(false),
		},
	}
	// Wrap in items array as expected by the test runner
	output := map[string]interface{}{
		"items": []interface{}{test},
	}
	out, err := yaml.Marshal(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding YAML: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(string(out))
}
func toItem[T any](resource interface{}) T {
	var item T
	if err := convertViaJSON(&item, resource); err != nil {
		panic(fmt.Sprintf("converting item: %v", err))
	}
	return item
}
func resourcesToItems[T any](resources ...interface{}) []T {
	items := make([]T, 0, len(resources))
	for _, res := range resources {
		items = append(items, toItem[T](res))
	}
	return items
}
func convertViaJSON(to, from any) error {
	bs, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, to)
}
