module code.cloudfoundry.org/quarks-job

require (
	code.cloudfoundry.org/quarks-utils v0.0.0-20200508141127-47307e498e12
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/pkg/errors v0.8.1
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	go.uber.org/zap v1.10.0
	gopkg.in/fsnotify.v1 v1.4.7
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
)

go 1.13
