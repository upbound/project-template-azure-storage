// Package main generates a CompositionTest
package main

import (
	"encoding/json"
	"fmt"
	"os"

	metav1 "dev.upbound.io/models/io/k8s/meta/v1"
	storagev1beta1 "dev.upbound.io/models/io/upbound/azure/storage/v1beta1"
	azv1beta1 "dev.upbound.io/models/io/upbound/azure/v1beta1"
	metav1alpha1 "dev.upbound.io/models/io/upbound/dev/meta/v1alpha1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

func main() {
	// Define assertions for the composed resources.
	// These assertions validate the correct structure and values of the resources
	// created by the compose-bucket function.
	assertResources := resourcesToItems[metav1alpha1.CompositionTestSpecAssertResourcesItem](
		// Assert ResourceGroup with correct location.
		&azv1beta1.ResourceGroup{
			APIVersion: ptr.To(azv1beta1.ResourceGroupAPIVersionazureUpboundIoV1Beta1),
			Kind:       ptr.To(azv1beta1.ResourceGroupKindResourceGroup),
			Spec: &azv1beta1.ResourceGroupSpec{
				ForProvider: &azv1beta1.ResourceGroupSpecForProvider{
					Location: ptr.To("eastus"),
				},
			},
		},
		// Assert Storage Account with correct configuration.
		&storagev1beta1.Account{
			APIVersion: ptr.To(storagev1beta1.AccountAPIVersionstorageAzureUpboundIoV1Beta1),
			Kind:       ptr.To(storagev1beta1.AccountKindAccount),
			Metadata: &metav1.ObjectMeta{
				Name: ptr.To("example"), // Name transformation (hyphens removed).
			},
			Spec: &storagev1beta1.AccountSpec{
				ForProvider: &storagev1beta1.AccountSpecForProvider{
					AccountTier:                     ptr.To("Standard"),
					AccountReplicationType:          ptr.To("LRS"),
					Location:                        ptr.To("eastus"),
					InfrastructureEncryptionEnabled: ptr.To(true),
					BlobProperties: &[]storagev1beta1.AccountSpecForProviderBlobPropertiesItem{{
						VersioningEnabled: ptr.To(true),
					}},
					ResourceGroupNameSelector: &storagev1beta1.AccountSpecForProviderResourceGroupNameSelector{
						MatchControllerRef: ptr.To(true),
					},
				},
			},
		},
		// Assert Storage Container with correct ACL mapping.
		&storagev1beta1.Container{
			APIVersion: ptr.To(storagev1beta1.ContainerAPIVersionstorageAzureUpboundIoV1Beta1),
			Kind:       ptr.To(storagev1beta1.ContainerKindContainer),
			Spec: &storagev1beta1.ContainerSpec{
				ForProvider: &storagev1beta1.ContainerSpecForProvider{
					ContainerAccessType: ptr.To("blob"), // ACL "public" maps to "blob".
					StorageAccountNameSelector: &storagev1beta1.ContainerSpecForProviderStorageAccountNameSelector{
						MatchControllerRef: ptr.To(true),
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
	// Wrap in items array as expected by the test runner.
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
