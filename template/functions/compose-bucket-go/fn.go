package main

import (
	"context"
	"encoding/json"
	"strings"

	"dev.upbound.io/models/com/example/platform/v1alpha1"
	metav1 "dev.upbound.io/models/io/k8s/meta/v1"
	storagev1beta1 "dev.upbound.io/models/io/upbound/azure/storage/v1beta1"
	azv1beta1 "dev.upbound.io/models/io/upbound/azure/v1beta1"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
	"k8s.io/utils/ptr"
)

// Function is your composition function.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())
	rsp := response.To(req, response.DefaultTTL)

	observedComposite, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get xr"))
		return rsp, nil
	}

	var xr v1alpha1.XStorageBucket
	if err := convertViaJSON(&xr, observedComposite.Resource); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot convert xr"))
		return rsp, nil
	}

	params := xr.Spec.Parameters
	if params.Location == nil || *params.Location == "" {
		response.Fatal(rsp, errors.New("missing location parameter"))
		return rsp, nil
	}

	// We'll collect our desired composed resources into this map, then convert
	// them to the SDK's types and set them in the response when we return.
	desiredComposed := make(map[resource.Name]any)
	defer func() {
		desiredComposedResources, err := request.GetDesiredComposedResources(req)
		if err != nil {
			response.Fatal(rsp, errors.Wrap(err, "cannot get desired resources"))
			return
		}

		for name, obj := range desiredComposed {
			c := composed.New()
			if err := convertViaJSON(c, obj); err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot convert %s to unstructured", name))
				return
			}
			desiredComposedResources[name] = &resource.DesiredComposed{Resource: c}
		}

		if err := response.SetDesiredComposedResources(rsp, desiredComposedResources); err != nil {
			response.Fatal(rsp, errors.Wrap(err, "cannot set desired resources"))
			return
		}
	}()

	// Determine container access type based on ACL
	containerAccessType := "private"
	if params.ACL != nil && *params.ACL == "public" {
		containerAccessType = "blob"
	}

	// Storage account names must be 3-24 character, lowercase alphanumeric
	// strings that are globally unique within Azure.
	accountName := ""
	if xr.Metadata != nil && xr.Metadata.Name != nil {
		accountName = strings.ReplaceAll(*xr.Metadata.Name, "-", "")
	}

	// Create ResourceGroup
	rg := &azv1beta1.ResourceGroup{
		APIVersion: ptr.To(azv1beta1.ResourceGroupAPIVersionazureUpboundIoV1Beta1),
		Kind:       ptr.To(azv1beta1.ResourceGroupKindResourceGroup),
		Spec: &azv1beta1.ResourceGroupSpec{
			ForProvider: &azv1beta1.ResourceGroupSpecForProvider{
				Location: params.Location,
			},
		},
	}
	desiredComposed["rg"] = rg

	// Create Storage Account
	matchControllerRef := true
	account := &storagev1beta1.Account{
		APIVersion: ptr.To(storagev1beta1.AccountAPIVersionstorageAzureUpboundIoV1Beta1),
		Kind:       ptr.To(storagev1beta1.AccountKindAccount),
		Metadata: &metav1.ObjectMeta{
			Name: &accountName,
		},
		Spec: &storagev1beta1.AccountSpec{
			ForProvider: &storagev1beta1.AccountSpecForProvider{
				AccountTier:                     ptr.To("Standard"),
				AccountReplicationType:          ptr.To("LRS"),
				Location:                        params.Location,
				InfrastructureEncryptionEnabled: ptr.To(true),
				BlobProperties: &[]storagev1beta1.AccountSpecForProviderBlobPropertiesItem{
					{
						VersioningEnabled: params.Versioning,
					},
				},
				ResourceGroupNameSelector: &storagev1beta1.AccountSpecForProviderResourceGroupNameSelector{
					MatchControllerRef: &matchControllerRef,
				},
			},
		},
	}
	desiredComposed["account"] = account

	// Create Storage Container
	container := &storagev1beta1.Container{
		APIVersion: ptr.To(storagev1beta1.ContainerAPIVersionstorageAzureUpboundIoV1Beta1),
		Kind:       ptr.To(storagev1beta1.ContainerKindContainer),
		Spec: &storagev1beta1.ContainerSpec{
			ForProvider: &storagev1beta1.ContainerSpecForProvider{
				ContainerAccessType: ptr.To(containerAccessType),
				StorageAccountNameSelector: &storagev1beta1.ContainerSpecForProviderStorageAccountNameSelector{
					MatchControllerRef: &matchControllerRef,
				},
			},
		},
	}
	desiredComposed["container"] = container

	return rsp, nil
}

func convertViaJSON(to, from any) error {
	bs, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, to)
}
