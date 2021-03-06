package cli_test

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("CLI", func() {
	act := func(arg ...string) (session *gexec.Session, err error) {
		cmd := exec.Command(cliPath, arg...)
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		return
	}

	BeforeEach(func() {
		os.Setenv("DOCKER_IMAGE_TAG", "v0.0.0")
	})

	Describe("help", func() {
		It("should show the help for server", func() {
			session, err := act("help")
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(Say(`Usage:`))
		})

		It("should show all available options for server", func() {
			session, err := act("help")
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(Say(`Flags:
      --apply-crd                         \(APPLY_CRD\) If true, apply CRDs on start \(default true\)
      --ctx-timeout int                   \(CTX_TIMEOUT\) context timeout for each k8s API request in seconds \(default 300\)
  -o, --docker-image-org string           \(DOCKER_IMAGE_ORG\) Dockerhub organization that provides the operator docker image \(default "cfcontainerization"\)
      --docker-image-pull-policy string   \(DOCKER_IMAGE_PULL_POLICY\) Image pull policy \(default "IfNotPresent"\)
  -r, --docker-image-repository string    \(DOCKER_IMAGE_REPOSITORY\) Dockerhub repository that provides the operator docker image \(default "quarks-job"\)
  -t, --docker-image-tag string           \(DOCKER_IMAGE_TAG\) Tag of the operator docker image \(default "\d+.\d+.\d+"\)
  -h, --help                              help for quarks-job
  -c, --kubeconfig string                 \(KUBECONFIG\) Path to a kubeconfig, not required in-cluster
  -l, --log-level string                  \(LOG_LEVEL\) Only print log messages from this level onward \(trace,debug,info,warn\) \(default "debug"\)
      --max-workers int                   \(MAX_WORKERS\) Maximum number of workers concurrently running the controller \(default 1\)
      --meltdown-duration int             \(MELTDOWN_DURATION\) Duration \(in seconds\) of the meltdown period, in which we postpone further reconciles for the same resource \(default 60\)
      --meltdown-requeue-after int        \(MELTDOWN_REQUEUE_AFTER\) Duration \(in seconds\) for which we delay the requeuing of the reconcile \(default 30\)
      --monitored-id string               \(MONITORED_ID\) only monitor namespaces with this id in their namespace label \(default "default"\)

`))
		})

		It("shows all available commands", func() {
			session, err := act("help")
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(Say(`Available Commands:
  help           Help about any command
  persist-output Persist a file into a kube secret
  version        Print the version number

`))
		})

		It("should show all available options for persist-output", func() {
			session, err := act("persist-output", "--help")
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(Say(`Flags:
  -h, --help               help for persist-output
      --namespace string   \(NAMESPACE\) namespace where persist output will run \(default "default"\)

`))
		})
	})

	Describe("default", func() {

		It("should start the server", func() {
			session, err := act()
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Err).Should(Say(`Starting quarks-job \d+\.\d+\.\d+, monitoring namespaces labeled with`))
			Eventually(session.Err).ShouldNot(Say(`Applying CRDs...`))
		})

		Context("when specifying monitored id for namespaces to monitor", func() {
			Context("via environment variables", func() {
				BeforeEach(func() {
					os.Setenv("MONITORED_ID", "env-test")
				})

				AfterEach(func() {
					os.Setenv("MONITORED_ID", "")
				})

				It("should start for that id", func() {
					session, err := act()
					Expect(err).ToNot(HaveOccurred())
					Eventually(session.Err).Should(Say(`Starting quarks-job \d+\.\d+\.\d+, monitoring namespaces labeled with 'env-test'`))
				})
			})

			Context("via using switches", func() {
				It("should start for namespace", func() {
					session, err := act("--monitored-id", "switch-test")
					Expect(err).ToNot(HaveOccurred())
					Eventually(session.Err).Should(Say(`Starting quarks-job \d+\.\d+\.\d+, monitoring namespaces labeled with 'switch-test'`))
				})
			})
		})

		Context("when enabling apply-crd", func() {
			Context("via environment variables", func() {
				BeforeEach(func() {
					os.Setenv("APPLY_CRD", "true")
				})

				AfterEach(func() {
					os.Setenv("APPLY_CRD", "")
				})

				It("should apply CRDs", func() {
					session, err := act()
					Expect(err).ToNot(HaveOccurred())
					Eventually(session.Err).Should(Say(`Applying CRDs...`))
				})
			})

			Context("via using switches", func() {
				It("should apply CRDs", func() {
					session, err := act("--apply-crd")
					Expect(err).ToNot(HaveOccurred())
					Eventually(session.Err).Should(Say(`Applying CRDs...`))
				})
			})
		})
	})

	Describe("version", func() {
		It("should show a semantic version number", func() {
			session, err := act("version")
			Expect(err).ToNot(HaveOccurred())
			Eventually(session.Out).Should(Say(`Quarks-Job Version: \d+.\d+.\d+`))
		})
	})
})
