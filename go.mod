module github.com/bmoussaud/buildpack-labels-operator

go 1.16

require (
	github.com/containers/image/v5 v5.13.2
	github.com/containers/ocicrypt v1.1.2 // indirect
	github.com/containers/storage v1.32.5 // indirect
	github.com/docker/docker v20.10.7+incompatible // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/imroc/req v0.3.0
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
	sigs.k8s.io/controller-runtime v0.8.3
)
