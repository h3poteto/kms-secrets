module github.com/h3poteto/kms-secrets

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.53
	github.com/go-logr/logr v0.4.0
	github.com/h3poteto/controller-klog v0.1.1
	github.com/onsi/ginkgo/v2 v2.1.4
	github.com/onsi/gomega v1.19.0
	gopkg.in/yaml.v3 v3.0.0
	k8s.io/api v0.22.8
	k8s.io/apimachinery v0.22.8
	k8s.io/client-go v0.22.8
	k8s.io/klog/v2 v2.9.0
	k8s.io/utils v0.0.0-20211116205334-6203023598ed
	sigs.k8s.io/controller-runtime v0.10.3
	sigs.k8s.io/yaml v1.2.0
)
