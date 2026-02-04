// Package main generates an E2ETest
package main

import (
	metav1 "dev.upbound.io/models/io/k8s/meta/v1"
	metav1alpha1 "dev.upbound.io/models/io/upbound/dev/meta/v1alpha1"
	"encoding/json"
	"fmt"
	"k8s.io/utils/ptr"
	"os"
	"sigs.k8s.io/yaml"
)

func main() {
	// Read Azure credentials from environment variable
	azureCreds := os.Getenv("UP_CLOUD_CREDENTIALS")
	if azureCreds == "" {
		fmt.Fprintf(os.Stderr, "Error: UP_CLOUD_CREDENTIALS environment variable not set\n")
		fmt.Fprintf(os.Stderr, "Please set it with your Azure service principal credentials:\n")
		fmt.Fprintf(os.Stderr, "  export UP_CLOUD_CREDENTIALS=$(cat azure-creds.json)\n")
		os.Exit(1)
	}

	// Define the StorageBucket XR to deploy for E2E testing
	// Note: XStorageBucket is cluster-scoped (no namespace)
	manifests := resourcesToItems[metav1alpha1.E2ETestSpecManifestsItem](
		map[string]interface{}{
			"apiVersion": "platform.example.com/v1alpha1",
			"kind":       "XStorageBucket",
			"metadata": map[string]interface{}{
				"name": "e2e-test-bucket",
			},
			"spec": map[string]interface{}{
				"parameters": map[string]interface{}{
					"location":   "eastus",
					"versioning": true,
					"acl":        "private",
				},
			},
		},
	)

	// Define extra resources: Secret and ProviderConfig
	extraResources := resourcesToItems[metav1alpha1.E2ETestSpecExtraResourcesItem](
		// Azure credentials Secret (created from environment variable)
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":      "azure-creds",
				"namespace": "upbound-system",
			},
			"stringData": map[string]interface{}{
				"credentials": azureCreds,
			},
		},
		// Azure ProviderConfig (references the Secret)
		map[string]interface{}{
			"apiVersion": "azure.upbound.io/v1beta1",
			"kind":       "ProviderConfig",
			"metadata": map[string]interface{}{
				"name": "default",
			},
			"spec": map[string]interface{}{
				"credentials": map[string]interface{}{
					"source": "Secret",
					"secretRef": map[string]interface{}{
						"namespace": "upbound-system",
						"name":      "azure-creds",
						"key":       "credentials",
					},
				},
			},
		},
	)

	test := metav1alpha1.E2ETest{
		APIVersion: ptr.To(metav1alpha1.E2ETestAPIVersionmetaDevUpboundIoV1Alpha1),
		Kind:       ptr.To(metav1alpha1.E2ETestKindE2ETest),
		Metadata: &metav1.ObjectMeta{
			Name: ptr.To("e2etest-bucket"),
		},
		Spec: &metav1alpha1.E2ETestSpec{
			DefaultConditions: &[]string{"Ready"},
			Manifests:         &manifests,
			ExtraResources:    &extraResources,
			TimeoutSeconds:    ptr.To(600), // 10 minutes for real Azure resources
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
