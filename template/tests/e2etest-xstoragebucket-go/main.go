// Package main generates an E2ETest
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"dev.upbound.io/models/com/example/platform/v1alpha1"
	metacorev1 "dev.upbound.io/models/io/k8s/core/meta/v1"
	corev1 "dev.upbound.io/models/io/k8s/core/v1"
	metav1 "dev.upbound.io/models/io/k8s/meta/v1"
	azv1beta1 "dev.upbound.io/models/io/upbound/azure/v1beta1"
	metav1alpha1 "dev.upbound.io/models/io/upbound/dev/meta/v1alpha1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

type e2eTestList struct {
	Items []metav1alpha1.E2ETest `json:"items"`
}

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
		&v1alpha1.XStorageBucket{
			APIVersion: ptr.To(v1alpha1.XStorageBucketAPIVersionplatformExampleComV1Alpha1),
			Kind:       ptr.To(v1alpha1.XStorageBucketKindXStorageBucket),
			Metadata: &metav1.ObjectMeta{
				Name: ptr.To("e2e-test-bucket"),
			},
			Spec: &v1alpha1.XStorageBucketSpec{
				Parameters: &v1alpha1.XStorageBucketSpecParameters{
					Location:   ptr.To("eastus"),
					Versioning: ptr.To(true),
					ACL:        ptr.To("private"),
				},
			},
		},
	)

	// Define extra resources: Secret and ProviderConfig
	extraResources := resourcesToItems[metav1alpha1.E2ETestSpecExtraResourcesItem](
		// Azure credentials Secret (created from environment variable)
		&corev1.Secret{
			APIVersion: ptr.To(corev1.SecretAPIVersionV1),
			Kind:       ptr.To(corev1.SecretKindSecret),
			Metadata: &metacorev1.ObjectMeta{
				Name:      ptr.To("azure-creds"),
				Namespace: ptr.To("upbound-system"),
			},
			StringData: &map[string]string{
				"credentials": azureCreds,
			},
		},
		// Azure ProviderConfig (references the Secret)
		&azv1beta1.ProviderConfig{
			APIVersion: ptr.To(azv1beta1.ProviderConfigAPIVersionazureUpboundIoV1Beta1),
			Kind:       ptr.To(azv1beta1.ProviderConfigKindProviderConfig),
			Metadata: &metav1.ObjectMeta{
				Name: ptr.To("default"),
			},
			Spec: &azv1beta1.ProviderConfigSpec{
				Credentials: &azv1beta1.ProviderConfigSpecCredentials{
					Source: ptr.To(azv1beta1.ProviderConfigSpecCredentialsSourceSecret),
					SecretRef: &azv1beta1.ProviderConfigSpecCredentialsSecretRef{
						Namespace: ptr.To("upbound-system"),
						Name:      ptr.To("azure-creds"),
						Key:       ptr.To("credentials"),
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
	output := e2eTestList{
		Items: []metav1alpha1.E2ETest{test},
	}
	out, err := yaml.Marshal(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding YAML: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(string(out))
}

func toItem[T any](resource any) T {
	var item T
	if err := convertViaJSON(&item, resource); err != nil {
		panic(fmt.Sprintf("converting item: %v", err))
	}
	return item
}

func resourcesToItems[T any](resources ...any) []T {
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
