package imagebuilder

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Report", func() {
	Describe("NewReportFromLogs", func() {
		logs := `Starting: prepare_image_build_report
==============================================================================
Task         : Python script
Description  : Run a Python file or inline script
Version      : 0.248.1
Author       : Microsoft Corporation
Help         : https://docs.microsoft.com/azure/devops/pipelines/tasks/utility/python-script
==============================================================================
/usr/bin/python /home/vsts/work/1/s/scripts/prepare_image_build_report.py --image-build-report-file /home/vsts/work/1/s/image-report.json --image-name ginkgo-test-image/ginkgo --sign-step-succeeded true --job-status Succeeded --image-build-report-file /home/vsts/work/_temp/generated-tags.json --images-to-sign=europe-docker.pkg.dev/kyma-project/prod/ginkgo-test-image/ginkgo:1.23.0-50049457 --images-to-sign=europe-docker.pkg.dev/kyma-project/prod/ginkgo-test-image/ginkgo:wartosc --images-to-sign=europe-docker.pkg.dev/kyma-project/prod/ginkgo-test-image/ginkgo:innytag --images-to-sign=europe-docker.pkg.dev/kyma-project/prod/ginkgo-test-image/ginkgo:v20250129-50049457 --images-to-sign=europe-docker.pkg.dev/kyma-project/prod/ginkgo-test-image/ginkgo:1.23.0
---IMAGE BUILD REPORT---
{
    "status": "Succeeded",
    "signed": true,
    "is_production": true,
    "image_spec": {
        "image_name": "ginkgo-test-image/ginkgo",
        "tags": [
            "1.23.0-50049457",
            "wartosc",
            "innytag",
            "v20250129-50049457",
            "1.23.0"
        ],
        "repository_path": "europe-docker.pkg.dev/kyma-project/prod/"
    }
}
---END OF IMAGE BUILD REPORT---

Finishing: prepare_image_build_report`
		expectedReport := &BuildReport{
			Status:       "Succeeded",
			IsSigned:     true,
			IsProduction: true,
			ImageSpec: ImageSpec{
				Name:           "ginkgo-test-image/ginkgo",
				Tags:           []string{"1.23.0-50049457", "wartosc", "innytag", "v20250129-50049457", "1.23.0"},
				RepositoryPath: "europe-docker.pkg.dev/kyma-project/prod/",
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
