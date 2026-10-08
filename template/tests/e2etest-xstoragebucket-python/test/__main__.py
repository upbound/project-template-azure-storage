import os
import base64

import yaml
from pydantic import BaseModel
from models.io.upbound.dev.meta.e2etest import v1alpha1 as e2etest
from models.io.k8s.apimachinery.pkg.apis.meta import v1 as k8s
from models.io.upbound.azure.providerconfig import v1beta1 as providerconfig
from models.com.example.platform.xstoragebucket import v1alpha1 as platformv1alpha1


class Secret(BaseModel):
    apiVersion: str = "v1"
    kind: str = "Secret"
    metadata: k8s.ObjectMeta
    type: str = "Opaque"
    data: dict[str, str] = {}

# Read Azure credentials from environment
azure_creds = os.environ.get("UP_AZURE_CREDS", "")
encoded_creds = base64.b64encode(azure_creds.encode()).decode()

# Define the XStorageBucket manifest
xstorage_bucket = platformv1alpha1.XStorageBucket(
    apiVersion="platform.example.com/v1alpha1",
    kind="XStorageBucket",
    metadata=k8s.ObjectMeta(
        name="uptest-bucket-xr-python"
    ),
    spec=platformv1alpha1.Spec(
        parameters=platformv1alpha1.Parameters(
            acl="public",
            location="eastus",
            versioning=True,
        )
    )
)

# Define the Azure provider config
provider_config = providerconfig.ProviderConfig(
    apiVersion="azure.upbound.io/v1beta1",
    kind="ProviderConfig",
    metadata=k8s.ObjectMeta(
        name="default"
    ),
    spec=providerconfig.Spec(
        credentials=providerconfig.Credentials(
            source="Secret",
            secretRef=providerconfig.SecretRef(
                key="credentials",
                name="azure-secret",
                namespace="crossplane-system",
            )
        )
    )
)

# Define the secret containing Azure credentials
azure_secret = Secret(
    apiVersion="v1",
    kind="Secret",
    metadata=k8s.ObjectMeta(
        name="azure-secret",
        namespace="crossplane-system"
    ),
    data={
        "credentials": encoded_creds
    }
)

test = e2etest.E2ETest(
    metadata=k8s.ObjectMeta(
        name="xstoragebucket-python",
    ),
    spec = e2etest.Spec(
        crossplane=e2etest.Crossplane(
            autoUpgrade=e2etest.AutoUpgrade(
                channel="Rapid",
            ),
        ),
        defaultConditions=[
            "Ready",
        ],
        manifests=[
            xstorage_bucket.model_dump(exclude_unset=True, by_alias=True),
        ],
        extraResources=[
            provider_config.model_dump(exclude_unset=True, by_alias=True),
            azure_secret.model_dump(exclude_unset=True, by_alias=True),
        ],
        skipDelete=False,
        timeoutSeconds=4500,
    )
)

# The test runner expects an "items" array, one entry per test.
output = {"items": [test.model_dump(by_alias=True, exclude_none=True)]}
print(yaml.dump(output))
