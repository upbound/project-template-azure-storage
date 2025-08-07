from .model.io.upbound.dev.meta.compositiontest import v1alpha1 as compositiontest
from .model.io.k8s.apimachinery.pkg.apis.meta import v1 as k8s
from .model.io.upbound.azure.resourcegroup import v1beta1 as rgv1beta1
from .model.io.upbound.azure.storage.account import v1beta1 as acctv1beta1
from .model.io.upbound.azure.storage.container import v1beta1 as contv1beta1
from .model.com.example.platform.xstoragebucket import v1alpha1 as platformv1alpha1

xStorageBucket = platformv1alpha1.XStorageBucket(
    apiVersion="platform.example.com/v1alpha1",
    kind="XStorageBucket",
    metadata=k8s.ObjectMeta(
        name="example"
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
    apiVersion="azure.upbound.io/v1beta1",
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
    apiVersion="storage.azure.upbound.io/v1beta1",
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
            blobProperties=[
                acctv1beta1.BlobProperty(
                    versioningEnabled=True,
                ),
            ],
            resourceGroupNameSelector=acctv1beta1.ResourceGroupNameSelector(
                matchControllerRef=True
            ),
        ),
    )
)

container = contv1beta1.Container(
    apiVersion="storage.azure.upbound.io/v1beta1",
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
        name="test-xstoragebucket-python",
    ),
    spec = compositiontest.Spec(
        assertResources=[
            xStorageBucket.model_dump(exclude_unset=True, by_alias=True),
            group.model_dump(exclude_unset=True, exclude={"spec": {"deletionPolicy", "managementPolicies"}}, by_alias=True),
            account.model_dump(exclude_unset=True, exclude={"spec": {"deletionPolicy", "managementPolicies"}}, by_alias=True),
            container.model_dump(exclude_unset=True, exclude={"spec": {"deletionPolicy", "managementPolicies"}}, by_alias=True),
        ],
        compositionPath="apis/xstoragebuckets/composition.yaml",
        xrPath="examples/xstoragebuckets/example.yaml",
        xrdPath="apis/xstoragebuckets/definition.yaml",
        timeoutSeconds=120,
        validate=False,
    )
)

# Export items for the test framework
items = [test.model_dump(exclude_unset=True, by_alias=True)]
