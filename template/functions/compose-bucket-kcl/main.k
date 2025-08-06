import models.io.upbound.azure.v1beta1 as azurev1beta1
import models.io.upbound.azure.storage.v1beta1 as storagev1beta1

oxr = option("params").oxr

containerAccessType = "blob" if oxr.spec.parameters.acl == "public" else "private"
accountName = oxr.metadata.name.replace("-", "")

_metadata = lambda name: str -> any {
  {
    annotations = {
      "krm.kcl.dev/composition-resource-name" = name
    }
  }
}

_items = [
    azurev1beta1.ResourceGroup{
        metadata = _metadata("rg")
        spec = {
            forProvider = {
                location = oxr.spec.parameters.location
            }
        }
    },
    storagev1beta1.Account{
        metadata = _metadata("account")
        metadata.name = accountName
        spec = {
            forProvider = {
                accountTier = "Standard"
                accountReplicationType = "LRS"
                location = oxr.spec.parameters.location
                blobProperties = [
                    {
                        versioningEnabled = oxr.spec.parameters.versioning
                    }
                ]
                infrastructureEncryptionEnabled = True
                resourceGroupNameSelector = {
                    matchControllerRef: True
                }
            }
        }
    },
    storagev1beta1.Container{
        metadata: _metadata("container")
        spec = {
            forProvider = {
                containerAccessType = containerAccessType
                storageAccountNameSelector = {
                    matchControllerRef: True
                }
            }
        }
    }
]
items = _items
