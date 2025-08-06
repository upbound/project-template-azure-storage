import models.io.upbound.azure.v1beta1 as azurev1beta1
import models.io.upbound.dev.meta.v1alpha1 as metav1alpha1
import models.com.example.platform.v1alpha1 as platformv1alpha1
import models.io.k8s.api.core.v1 as corev1
import base64
import file

azure_creds = file.read_env("UP_AZURE_CREDS")

_items = [
    metav1alpha1.E2ETest{
        metadata.name: "xstoragebucket-kcl"
        spec= {
            crossplane.autoUpgrade.channel: "Rapid"
            defaultConditions: ["Ready"]
            manifests: [
                platformv1alpha1.XStorageBucket{
                    metadata.name = "uptest-bucket-xr-kcl"
                    spec.parameters = {
                        acl = "public"
                        location = "eastus"
                        versioning: True
                    }
                }
            ]
            extraResources: [
                azurev1beta1.ProviderConfig{
                    metadata.name = "default"
                    spec: {
                        credentials: {
                            source: "Secret"
                            secretRef: {
                                key: "credentials"
                                name: "azure-secret"
                                namespace: "crossplane-system"
                            }
                        }
                    }
                }
                corev1.Secret{
                    metadata: {
                        name: "azure-secret"
                        namespace: "crossplane-system"
                    }
                    data={
                        "credentials": base64.encode(azure_creds)
                    }
                }

            ]
            skipDelete: False
            timeoutSeconds: 4500
        }
    }
]

items= _items