from crossplane.function import resource
from crossplane.function.proto.v1 import run_function_pb2 as fnv1

from .model.io.k8s.apimachinery.pkg.apis.meta import v1 as metav1
from .model.io.upbound.azure.resourcegroup import v1beta1 as rgv1beta1
from .model.io.upbound.azure.storage.account import v1beta1 as acctv1beta1
from .model.io.upbound.azure.storage.container import v1beta1 as contv1beta1
from .model.com.example.platform.xstoragebucket import v1alpha1


def compose(req: fnv1.RunFunctionRequest, rsp: fnv1.RunFunctionResponse):
    observed_xr = v1alpha1.XStorageBucket(**req.observed.composite.resource)
    params = observed_xr.spec.parameters

    # Create the resource group
    desired_group = rgv1beta1.ResourceGroup(
        spec=rgv1beta1.Spec(
            forProvider=rgv1beta1.ForProvider(
                location=params.location,
            ),
        ),
    )
    resource.update(rsp.desired.resources["rg"], desired_group)

    # Storage account names must be 3-24 character, lowercase alphanumeric
    # strings that are globally unique within Azure. We try to generate a valid
    # one automatically by deriving it from the XR name, which should always be
    # alphanumeric, lowercase, and separated by hyphens.
    account_external_name = observed_xr.metadata.name.replace("-", "")  # type: ignore  # Name is an optional field, but it'll always be set.

    # Create the storage account
    desired_acct = acctv1beta1.Account(
        metadata=metav1.ObjectMeta(
            name=account_external_name,
        ),
        spec=acctv1beta1.Spec(
            forProvider=acctv1beta1.ForProvider(
                accountTier="Standard",
                accountReplicationType="LRS",
                location=params.location,
                infrastructureEncryptionEnabled=True,
                blobProperties=[
                    acctv1beta1.BlobProperty(
                        versioningEnabled=params.versioning,
                    ),
                ],
                resourceGroupNameSelector=acctv1beta1.ResourceGroupNameSelector(
                    matchControllerRef=True
                ),
            ),
        ),
    )
    resource.update(rsp.desired.resources["account"], desired_acct)

    # Create the storage container
    desired_cont = contv1beta1.Container(
        metadata=metav1.ObjectMeta(
        ),
        spec=contv1beta1.Spec(
            forProvider=contv1beta1.ForProvider(
                containerAccessType="blob" if params.acl == "public" else "private",
                storageAccountNameSelector=contv1beta1.StorageAccountNameSelector(
                    matchControllerRef=True
                ),
            ),
        ),
    )
    resource.update(rsp.desired.resources["container"], desired_cont)
