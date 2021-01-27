module github.com/h3poteto/kms-secrets

go 1.15

require (
	github.com/aws/aws-sdk-go v1.36.32
	github.com/go-logr/logr v0.3.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog v1.0.0 // indirect
	sigs.k8s.io/controller-runtime v0.8.1
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
)
