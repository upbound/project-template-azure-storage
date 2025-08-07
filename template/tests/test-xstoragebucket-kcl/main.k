import models.com.example.platform.v1alpha1 as platformv1alpha1
import models.io.upbound.azure.v1beta1 as azurev1beta1
import models.io.upbound.azure.storage.v1beta1 as storagev1beta1
import models.io.upbound.dev.meta.v1alpha1 as metav1alpha1

# It's a best practice for composition functions to return only fields whose
# values they care about, since they become the owner of any field they
# return. Allow for fields with defaults to be omitted by the composition
# function by clearing them in our expected resources.
_stripDefaults = lambda obj: any -> any {
    obj | {
        spec.deletionPolicy = Undefined
        spec.managementPolicies = Undefined
    }
}

_items = [
    metav1alpha1.CompositionTest{
        metadata.name="test-xstoragebucket-kcl"
        spec= {
            assertResources: [
                platformv1alpha1.XStorageBucket{
                    metadata.name: "example"
                    spec: {
                        parameters: {
                            acl: "public"
                            location: "eastus"
                            versioning: True
                        }
                    }
                }
                _stripDefaults(azurev1beta1.ResourceGroup{
                    metadata: {
                        annotations: {
                            "crossplane.io/composition-resource-name": "rg"
                        }
                    }
                    spec.forProvider: {
                        location: "eastus"
                    }
                })
                _stripDefaults(storagev1beta1.Account{
                    metadata: {
                        annotations: {
                            "crossplane.io/composition-resource-name": "account"
                        }
                        name: "example"
                    }
                    spec.forProvider: {
                        accountTier: "Standard"
                        accountReplicationType: "LRS"
                        location: "eastus"
                        blobProperties = [
                            {
                                versioningEnabled = True
                            }
                        ]
                        infrastructureEncryptionEnabled = True
                        resourceGroupNameSelector = {
                            matchControllerRef: True
                        }
                    }
                })
                _stripDefaults(storagev1beta1.Container{
                    metadata: {
                        annotations: {
                            "crossplane.io/composition-resource-name": "container"
                        }
                    }
                    spec.forProvider: {
                        containerAccessType: "blob"
                        storageAccountNameSelector: {
                            matchControllerRef: True
                        }
                    }
                })
            ]
            compositionPath: "apis/xstoragebuckets/composition.yaml"
            xrPath: "examples/xstoragebuckets/example.yaml"
            xrdPath: "apis/xstoragebuckets/definition.yaml"
            timeoutSeconds: 120
            validate: False
        }
    }
]

items = _items
