module github.com/aws-controllers-k8s/eks-controller

go 1.14

replace github.com/aws-controllers-k8s/runtime => ../runtime

require (
	github.com/aws-controllers-k8s/runtime v0.16.0
	github.com/aws/aws-sdk-go v1.38.67
	github.com/go-logr/logr v1.2.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/objx v0.2.0 // indirect
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	sigs.k8s.io/controller-runtime v0.11.0
)
