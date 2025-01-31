package imagebuilder

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Report", func() {
	Describe("NewReportFromLogs", func() {
		logs := `2025-01-31T08:32:23.5327056Z ##[section]Starting: prepare_image_build_report
2025-01-31T08:32:23.5434336Z ==============================================================================
2025-01-31T08:32:23.5434499Z Task         : Python script
2025-01-31T08:32:23.5434594Z Description  : Run a Python file or inline script
2025-01-31T08:32:23.5434703Z Version      : 0.248.1
2025-01-31T08:32:23.5434803Z Author       : Microsoft Corporation
2025-01-31T08:32:23.5434910Z Help         : https://docs.microsoft.com/azure/devops/pipelines/tasks/utility/python-script
2025-01-31T08:32:23.5435059Z ==============================================================================
2025-01-31T08:32:23.6965198Z [command]/opt/hostedtoolcache/Python/3.13.1/x64/bin/python /home/vsts/work/1/s/scripts/prepare_image_build_report.py --image-name github-tools-sap/conduit-cli --image-build-succeeded true --sign-step-succeeded $(sign_images.signing_success) --job-status Succeeded --images-to-sign=europe-docker.pkg.dev/kyma-project/dev/github-tools-sap/conduit-cli:PR-477
2025-01-31T08:32:23.7344251Z ---IMAGE BUILD REPORT---
2025-01-31T08:32:23.7345746Z {
2025-01-31T08:32:23.7346062Z     "status": "Succeeded",
2025-01-31T08:32:23.7357582Z     "pushed": true,
2025-01-31T08:32:23.7358184Z     "signed": false,
2025-01-31T08:32:23.7358759Z     "is_production": false,
2025-01-31T08:32:23.7359525Z     "image_spec": {
2025-01-31T08:32:23.7360295Z         "image_name": "github-tools-sap/conduit-cli",
2025-01-31T08:32:23.7360618Z         "tags": [
2025-01-31T08:32:23.7361207Z             "PR-477"
2025-01-31T08:32:23.7361687Z         ],
2025-01-31T08:32:23.7362370Z         "repository_path": "europe-docker.pkg.dev/kyma-project/dev/"
2025-01-31T08:32:23.7362690Z     }
2025-01-31T08:32:23.7363276Z }
2025-01-31T08:32:23.7363903Z ---END OF IMAGE BUILD REPORT---
2025-01-31T08:32:23.7416532Z 
2025-01-31T08:32:23.7530550Z ##[section]Finishing: prepare_image_build_report`
		expectedReport := &BuildReport{
			Status:       "Succeeded",
			IsPushed:     true,
			IsSigned:     false,
			IsProduction: false,
			ImageSpec: ImageSpec{
				Name:           "github-tools-sap/conduit-cli",
				Tags:           []string{"PR-477"},
				RepositoryPath: "europe-docker.pkg.dev/kyma-project/dev/",
			},
		}

		It("parses the image build report", func() {
			actual, err := NewBuildReportFromLogs(logs)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(Equal(expectedReport))
		})
	})

	Describe("GetImages", func() {
		report := &BuildReport{
			ImageSpec: ImageSpec{
				Name:           "ginkgo-test-image/ginkgo",
				Tags:           []string{"1.23.0-50049457", "wartosc", "innytag", "v20250129-50049457", "1.23.0"},
				RepositoryPath: "europe-docker.pkg.dev/kyma-project/prod/",
			},
		}

		It("returns the list of images", func() {
			expectedImages := []string{
				"europe-docker.pkg.dev/kyma-project/prod/ginkgo-test-image/ginkgo:1.23.0-50049457",
				"europe-docker.pkg.dev/kyma-project/prod/ginkgo-test-image/ginkgo:wartosc",
				"europe-docker.pkg.dev/kyma-project/prod/ginkgo-test-image/ginkgo:innytag",
				"europe-docker.pkg.dev/kyma-project/prod/ginkgo-test-image/ginkgo:v20250129-50049457",
				"europe-docker.pkg.dev/kyma-project/prod/ginkgo-test-image/ginkgo:1.23.0",
			}

			Expect(report.GetImages()).To(Equal(expectedImages))
		})

		It("returns an empty list if there are no tags", func() {
			report.ImageSpec.Tags = []string{}
			Expect(report.GetImages()).To(BeEmpty())
		})

		It("returns an empty list if build report is nil", func() {
			var nilReport *BuildReport
			Expect(nilReport.GetImages()).To(BeEmpty())
		})
	})
})
