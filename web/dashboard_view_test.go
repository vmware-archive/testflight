package web_test

import (
	"github.com/concourse/atc"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"

	"encoding/json"
	"fmt"
	"time"

	"github.com/concourse/go-concourse/concourse"
	"github.com/concourse/testflight/helpers"
)

var _ = FDescribe("DashboardView", func() {
	var (
		otherTeam           concourse.Team
		otherTeamName       string
		privatePipelineName string
		publicPipelineName  string
		loadingTimeout      time.Duration
	)

	BeforeEach(func() {
		loadingTimeout = 10 * time.Second

		otherTeamName = fmt.Sprintf("test-team-%d", GinkgoParallelNode())
		_, created, _, err := client.Team(otherTeamName).CreateOrUpdate(atc.Team{
			Name: otherTeamName,
			Auth: make(map[string]*json.RawMessage),
		})
		Expect(created).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())

		otherClient := helpers.ConcourseClient(atcURL, otherTeamName)
		otherTeam = otherClient.Team(otherTeamName)

		config := atc.Config{
			Jobs: []atc.JobConfig{
				{Name: "some-job-name"},
			},
		}

		byteConfig, err := yaml.Marshal(config)
		Expect(err).NotTo(HaveOccurred())

		_, _, _, err = team.CreateOrUpdatePipelineConfig(pipelineName, "0", byteConfig)
		Expect(err).NotTo(HaveOccurred())

		privatePipelineName = pipelineName + "-private"
		_, _, _, err = otherTeam.CreateOrUpdatePipelineConfig(privatePipelineName, "0", byteConfig)
		Expect(err).NotTo(HaveOccurred())

		publicPipelineName = pipelineName + "-public"
		_, _, _, err = otherTeam.CreateOrUpdatePipelineConfig(publicPipelineName, "0", byteConfig)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_, err := otherTeam.DeletePipeline(privatePipelineName)
		Expect(err).ToNot(HaveOccurred())

		_, err = otherTeam.DeletePipeline(publicPipelineName)
		Expect(err).ToNot(HaveOccurred())

		err = otherTeam.DestroyTeam(otherTeamName)
		Expect(err).ToNot(HaveOccurred())
	})

	It("shows all pipelines from the authenticated team and public pipelines from other teams", func() {
		dashboardUrl := atcRoute("/dashboard")
		Expect(page.Navigate(dashboardUrl)).To(Succeed())
		Eventually(page, loadingTimeout).Should(HaveURL(dashboardUrl))

		teamNode := page.Find(fmt.Sprintf(".team:contains('%s')", "main"))
		Expect(teamNode).To(HaveText(pipelineName))

		otherTeamNode := page.Find(fmt.Sprintf(".team:contains('%s')", otherTeamName))
		Expect(otherTeamNode).To(HaveText(publicPipelineName))
		Expect(otherTeamNode).ToNot(HaveText(privatePipelineName))
	})
})
