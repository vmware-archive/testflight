package pipelines_test

import (
	"github.com/concourse/testflight/gitserver"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("image resource caching", func() {
	var (
		originGitServer *gitserver.Server
	)

	BeforeEach(func() {
		originGitServer = gitserver.Start(client)
	})

	AfterEach(func() {
		originGitServer.Stop()
	})

	It("gets the image resource from a cached resource with the same params", func() {
		originGitServer.CommitResource()
		originGitServer.CommitFileToBranch("initial", "initial", "trigger")

		hash, err := uuid.NewV4()
		Expect(err).ToNot(HaveOccurred())

		By("configuring the pipeline")
		flyHelper.ConfigurePipeline(
			pipelineName,
			"-c", "fixtures/image-resource-with-params.yml",
			"-v", "origin-git-server="+originGitServer.URI(),
			"-v", "hash="+hash.String(),
		)

		By("triggering the resource get with params")
		watch := flyHelper.TriggerJob(pipelineName, "with-params")
		<-watch.Exited
		Expect(watch).To(gbytes.Say("Cloning"))

		By("triggering the task with image resource with params")
		watch = flyHelper.TriggerJob(pipelineName, "image-resource-with-params")
		<-watch.Exited
		Expect(watch).ToNot(gbytes.Say("Cloning"))

		By("triggering the task with image resource without params")
		watch = flyHelper.TriggerJob(pipelineName, "image-resource-without-params")
		<-watch.Exited
		Expect(watch).To(gbytes.Say("Cloning"))
	})
})
