package pipelines_test

import (
	"fmt"

	"github.com/concourse/testflight/gitserver"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Resource config versions", func() {
	var originGitServer *gitserver.Server

	BeforeEach(func() {
		originGitServer = gitserver.Start(client)
		originGitServer.CommitResource()
	})

	AfterEach(func() {
		originGitServer.Stop()
	})

	// This test is for a case where the build inputs and outputs will not be invalidated if the resource config id field on the resource
	// gets updated due to a new version of the custom resource type that it is using.
	Describe("build inputs and outputs are not affected by a change in resource config id", func() {
		It("will run both jobs only once even with a new custom resource type version", func() {
			hash, err := uuid.NewV4()
			Expect(err).ToNot(HaveOccurred())

			flyHelper.ConfigurePipeline(
				pipelineName,
				"-c", "fixtures/resource-type-versions.yml",
				"-v", "origin-git-server="+originGitServer.URI(),
				"-v", "hash="+hash.String(),
			)

			By("Waiting for a new build when the pipeline is created")
			watch := flyHelper.Watch(pipelineName, "initial-job")
			<-watch.Exited
			Expect(watch).To(gbytes.Say("succeeded"))
			Expect(watch).To(gexec.Exit(0))

			By("Committing to change the custom type")
			originGitServer.CommitFileToBranch("new-contents", "rootfs/some-file", "master")

			By("Checking the custom resource type")
			checkResourceType := flyHelper.CheckResourceType("-r", fmt.Sprintf("%s/custom", pipelineName))
			<-checkResourceType.Exited
			Expect(checkResourceType.ExitCode()).To(Equal(0))
			Expect(checkResourceType).To(gbytes.Say("checked 'custom'"))

			By("Checking the resource using the custom type")
			checkResource := flyHelper.CheckResource("-r", fmt.Sprintf("%s/some-resource", pipelineName))
			<-checkResource.Exited
			Expect(checkResource.ExitCode()).To(Equal(0))
			Expect(checkResource).To(gbytes.Say("checked 'some-resource'"))

			By("Triggering a job using the custom type")
			watch = flyHelper.TriggerJob(pipelineName, "passed-job")
			<-watch.Exited
			Expect(watch).To(gbytes.Say("succeeded"))
			Expect(watch).To(gexec.Exit(0))

			By("Using the version  of 'some-resource' consumed upstream")
			builds := flyHelper.Builds(pipelineName, "initial-job")
			Expect(builds).To(HaveLen(1))
		})
	})
})
