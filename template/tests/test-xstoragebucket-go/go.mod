module github.com/upbound/project-template-azure-storage/tests/test-bucket

go 1.24.9

require (
	dev.upbound.io/models v0.0.0
	k8s.io/utils v0.0.0-20241104163129-6fe5fd82f078
	sigs.k8s.io/yaml v1.4.0
)

replace dev.upbound.io/models => ../../.up/go/models
