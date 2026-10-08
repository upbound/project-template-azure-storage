import yaml
from models.io.upbound.dev.meta.compositiontest import v1alpha1 as compositiontest
from models.io.k8s.apimachinery.pkg.apis.meta import v1 as k8s
from models.io.upbound.m.azure.resourcegroup import v1beta1 as rgv1beta1
from models.io.upbound.m.azure.storage.account import v1beta1 as acctv1beta1
from models.io.upbound.m.azure.storage.container import v1beta1 as contv1beta1
from models.com.example.platform.storagebucket import v1alpha1 as platformv1alpha1

storageBucket = platformv1alpha1.StorageBucket(
    apiVersion="platform.example.com/v1alpha1",
    kind="StorageBucket",
    metadata=k8s.ObjectMeta(
        name="example",
        namespace="default",
    ),
    spec = platformv1alpha1.Spec(
        parameters = platformv1alpha1.Parameters(
            acl="public",
            location="eastus",
            versioning=True,
        ),
    ),
)

group = rgv1beta1.ResourceGroup(
    apiVersion="azure.m.upbound.io/v1beta1",
    kind="ResourceGroup",
    metadata=k8s.ObjectMeta(
        annotations={
            "crossplane.io/composition-resource-name": "rg"
        }
    ),
    spec=rgv1beta1.Spec(
        forProvider=rgv1beta1.ForProvider(
            location="eastus",
        )
    )
)

account = acctv1beta1.Account(
    apiVersion="storage.azure.m.upbound.io/v1beta1",
    kind="Account",
    metadata=k8s.ObjectMeta(
        name="example",
        annotations={
            "crossplane.io/composition-resource-name": "account"
        }
    ),
    spec=acctv1beta1.Spec(
        forProvider=acctv1beta1.ForProvider(
            accountTier="Standard",
            accountReplicationType="LRS",
            location="eastus",
            infrastructureEncryptionEnabled=True,
            blobProperties=acctv1beta1.BlobProperties(
                versioningEnabled=True,
            ),
            resourceGroupNameSelector=acctv1beta1.ResourceGroupNameSelector(
                matchControllerRef=True
            ),
        ),
    )
)

container = contv1beta1.Container(
    apiVersion="storage.azure.m.upbound.io/v1beta1",
    kind="Container",
    metadata=k8s.ObjectMeta(
        annotations={
            "crossplane.io/composition-resource-name": "container"
        }
    ),
    spec=contv1beta1.Spec(
        forProvider=contv1beta1.ForProvider(
            containerAccessType="blob",
            storageAccountNameSelector=contv1beta1.StorageAccountNameSelector(
                matchControllerRef=True
            ),
        )
    )
)

test = compositiontest.CompositionTest(
    metadata=k8s.ObjectMeta(
        name="test-storagebucket-python",
    ),
    spec = compositiontest.Spec(
        assertResources=[
            storageBucket.model_dump(exclude_unset=True, by_alias=True),
            group.model_dump(exclude_unset=True, exclude={"spec": {"managementPolicies"}}, by_alias=True),
            account.model_dump(exclude_unset=True, exclude={"spec": {"managementPolicies"}}, by_alias=True),
            container.model_dump(exclude_unset=True, exclude={"spec": {"managementPolicies"}}, by_alias=True),
        ],
        compositionPath="apis/storagebuckets/composition.yaml",
        xrPath="examples/storagebuckets/example.yaml",
        xrdPath="apis/storagebuckets/definition.yaml",
        timeoutSeconds=120,
        validate=False,
    )
)

# The test runner expects an "items" array, one entry per test.
output = {"items": [test.model_dump(by_alias=True, exclude_none=True)]}
print(yaml.dump(output))
