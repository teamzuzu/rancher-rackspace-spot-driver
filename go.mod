module github.com/teamzuzu/rancher-rackspace-spot-driver

go 1.24

require (
	github.com/rackspace-spot/spot-go-sdk v0.2.0
	github.com/rancher/kontainer-engine v0.0.4-dev.0.20210625182816-1a4f4e73a324
	github.com/sirupsen/logrus v1.9.3
	k8s.io/api v0.27.4
	k8s.io/apimachinery v0.27.4
	k8s.io/client-go v0.27.4
)

// kontainer-engine (via rancher/rke) transitively pulls k8s.io/client-go v12.0.0+incompatible,
// which references alpha API packages removed in k8s.io/api >= v0.28. Pinning all three k8s
// modules prevents go mod tidy from trying to resolve those packages against a newer version.
replace (
	k8s.io/api => k8s.io/api v0.27.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.27.4
	k8s.io/client-go => k8s.io/client-go v0.27.4
)
