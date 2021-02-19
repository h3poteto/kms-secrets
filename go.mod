module github.com/h3poteto/kms-secrets

go 1.15

require (
	github.com/aws/aws-sdk-go v1.37.1
	github.com/go-logr/logr v0.4.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	k8s.io/klog/v2 v2.5.0
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.1
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
	sigs.k8s.io/yaml v1.2.0
)
