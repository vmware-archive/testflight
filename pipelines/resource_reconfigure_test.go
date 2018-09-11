package pipelines_test

import (
	"github.com/concourse/testflight/gitserver"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Reconfiguring a resource", func() {
	var originGitServer *gitserver.Server

	BeforeEach(func() {
		originGitServer = gitserver.Start(client)
	})

	AfterEach(func() {
		originGitServer.Stop()
	})

	It("creates a new check container with the updated configuration", func() {
		hash, err := uuid.NewV4()
		Expect(err).ToNot(HaveOccurred())

		flyHelper.ConfigurePipeline(
			pipelineName,
			"-c", "fixtures/simple-trigger.yml",
			"-v", "origin-git-server="+originGitServer.URI(),
			"-v", "hash="+hash.String(),
		)

		guid1 := originGitServer.Commit()
		watch := flyHelper.Watch(pipelineName, "some-passing-job", "1")
		Eventually(watch).Should(gbytes.Say(guid1))

		flyHelper.ReconfigurePipeline(
			pipelineName,
			"-c", "fixtures/simple-trigger-reconfigured.yml",
			"-v", "origin-git-server="+originGitServer.URI(),
		)

		guid2 := originGitServer.CommitOnBranch("other")
		watch = flyHelper.Watch(pipelineName, "some-passing-job", "2")
		Eventually(watch).Should(gbytes.Say(guid2))
	})
})
