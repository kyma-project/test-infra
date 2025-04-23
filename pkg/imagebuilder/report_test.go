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
2025-01-31T08:32:23.7358759Z     "tags": [
2025-01-31T08:32:23.7359525Z         "PR-477"
2025-01-31T08:32:23.7360295Z     ],
2025-01-31T08:32:23.7360618Z     "repository_path": "europe-docker.pkg.dev/kyma-project/dev/",
2025-01-31T08:32:23.7361207Z     "image_name": "github-tools-sap/conduit-cli",
2025-01-31T08:32:23.7363009Z     "images_list": [
2025-01-31T08:32:23.7363276Z         "europe-docker.pkg.dev/kyma-project/dev/github-tools-sap/conduit-cli:PR-477"
2025-01-31T08:32:23.7363276Z     ],
2025-01-31T08:32:23.7363276Z     "digest": "sha256:215151561",
2025-01-31T08:32:23.7363276Z     "architecture": [
2025-01-31T08:32:23.7363276Z         "linux/amd64"
2025-01-31T08:32:23.7363276Z     ]
2025-01-31T08:32:23.7363276Z }
2025-01-31T08:32:23.7363903Z ---END OF IMAGE BUILD REPORT---
2025-01-31T08:32:23.7416532Z 
2025-01-31T08:32:23.7530550Z ##[section]Finishing: prepare_image_build_report`
		expectedReport := &BuildReport{
			Status:       "Succeeded",
			IsPushed:     true,
			IsSigned:     false,
			Images:       []string{"europe-docker.pkg.dev/kyma-project/dev/github-tools-sap/conduit-cli:PR-477"},
			Digest:       "sha256:215151561",
			Name:         "github-tools-sap/conduit-cli",
			Tags:         []string{"PR-477"},
			RegistryURL:  "europe-docker.pkg.dev/kyma-project/dev/",
			Architecture: []string{"linux/amd64"},
		}

		It("parses the image build report", func() {
			actual, err := NewBuildReportFromLogs(logs)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(Equal(expectedReport))
		})

		It("returns an error if the log does not contain the image build report", func() {
			logs := `2025-01-31T08:32:23.5327056Z ##[section]Starting: prepare_image_build_report`

			_, err := NewBuildReportFromLogs(logs)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("WriteReportToFile", func() {
		report := &BuildReport{
			Status:      "Succeeded",
			IsPushed:    true,
			IsSigned:    false,
			Images:      []string{"europe-docker.pkg.dev/kyma-project/dev/github-tools-sap/conduit-cli:PR-477"},
			Digest:      "sha256:215151561",
			Name:        "github-tools-sap/conduit-cli",
			Tags:        []string{"PR-477"},
			RegistryURL: "europe-docker.pkg.dev/kyma-project/dev/",
		}

		It("writes the report to a file", func() {
			path := "/tmp/report.json"
			err := WriteReportToFile(report, path)
			Expect(err).ToNot(HaveOccurred())

			Expect(path).To(BeAnExistingFile())
		})
	})
})
