module github.com/h3poteto/kms-secrets

go 1.15

require (
	github.com/aws/aws-sdk-go v1.38.53
	github.com/go-logr/logr v0.4.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/h3poteto/controller-klog v0.1.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.10.5
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog/v2 v2.9.0
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/yaml v1.2.0
)
