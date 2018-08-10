package pipelines_test

import (
	"fmt"
	"time"

	"github.com/concourse/testflight/gitserver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var checkContents = `#!/bin/sh
sleep 10000
`
var defaultCheckTimeout = 1 * time.Minute

var _ = FDescribe("resource check time out", func() {
	var originGitServer *gitserver.Server
	var configPath string
	var checkStartTime time.Time
	var checkEndTime time.Time
	var checkSession *gexec.Session

	BeforeEach(func() {
		originGitServer = gitserver.Start(client)
		originGitServer.CommitResource()
		originGitServer.CommitFileToBranch(checkContents, "rootfs/opt/resource/check", "master")
		checkStartTime = time.Now()
	})

	JustBeforeEach(func() {
		flyHelper.ConfigurePipeline(
			pipelineName,
			"-c", configPath,
			"-v", "origin-git-server="+originGitServer.URI(),
			"-y", "privileged=true",
		)

		checkSession = flyHelper.CheckResource("-r", fmt.Sprintf("%s/my-resource", pipelineName))
		<-checkSession.Exited
		checkEndTime = time.Now()
	})

	Context("when only default time out is set", func() {
		BeforeEach(func() {
			configPath = "fixtures/resource-check-timeouts-default.yml"
		})

		It("times out with global default duration when checking", func() {
			Expect(checkSession.Err).To(gbytes.Say("check-timed-out"))
			Expect(checkSession).To(gexec.Exit(1))
			Expect(checkEndTime.Before(checkStartTime.Add(defaultCheckTimeout + 1*time.Minute))).To(BeTrue())
		})
	})

	Context("when per resource check time out is set", func() {
		BeforeEach(func() {
			configPath = "fixtures/resource-check-timeouts-custom.yml"
		})

		It("times out with defined duration(1 min) in resource config when checking", func() {
			Expect(checkSession.Err).To(gbytes.Say("check-timed-out"))
			Expect(checkSession).To(gexec.Exit(1))
			Expect(checkEndTime.Before(checkStartTime.Add(2 * time.Minute))).To(BeTrue())
		})
	})
})
