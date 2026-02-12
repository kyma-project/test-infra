package imagebuilder

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Report", func() {
	Describe("NewReportFromLogs", func() {
		logs := `2025-04-27T10:41:56.2511247Z ##[section]Starting: print_image_build_report
2025-04-27T10:41:56.2516763Z ==============================================================================
2025-04-27T10:41:56.2517121Z Task         : Bash
2025-04-27T10:41:56.2517335Z Description  : Run a Bash script on macOS, Linux, or Windows
2025-04-27T10:41:56.2517496Z Version      : 3.250.1
2025-04-27T10:41:56.2517599Z Author       : Microsoft Corporation
2025-04-27T10:41:56.2517740Z Help         : https://docs.microsoft.com/azure/devops/pipelines/tasks/utility/bash
2025-04-27T10:41:56.2517900Z ==============================================================================
2025-04-27T10:41:56.4166633Z Generating script.
2025-04-27T10:41:56.4179864Z ========================== Starting Command Output ===========================
2025-04-27T10:41:56.4189785Z [command]/usr/bin/bash /home/vsts/work/_temp/4968a323-f629-41c2-90a2-53caa1477a37.sh
2025-04-27T10:41:56.4259451Z [#debug] Build Report: {status: Succeeded, pushed: true, signed: false, image_name: cors-proxy, tags: [PR-12975], repository_path: europe-docker.pkg.dev/kyma-project/dev, images_list: [europe-docker.pkg.dev/kyma-project/dev/cors-proxy:PR-12975], digest: sha256:3197820c25f93113f22a6d90d6dbcf70e1d71ae528c3c0b1542e9604bdfa9d83, architectures: [linux/arm64, linux/amd64]}
2025-04-27T10:41:56.4260338Z ---IMAGE BUILD REPORT---
2025-04-27T10:41:56.4342736Z {
2025-04-27T10:41:56.4346228Z   "status": "Succeeded",
2025-04-27T10:41:56.4346484Z   "pushed": true,
2025-04-27T10:41:56.4346725Z   "signed": false,
2025-04-27T10:41:56.4347489Z   "image_name": "cors-proxy",
2025-04-27T10:41:56.4347733Z   "tags": [
2025-04-27T10:41:56.4348026Z     "PR-12975"
2025-04-27T10:41:56.4348242Z   ],
2025-04-27T10:41:56.4348636Z   "repository_path": "europe-docker.pkg.dev/kyma-project/dev",
2025-04-27T10:41:56.4348920Z   "images_list": [
2025-04-27T10:41:56.4349601Z     "europe-docker.pkg.dev/kyma-project/dev/cors-proxy:PR-12975"
2025-04-27T10:41:56.4349903Z   ],
2025-04-27T10:41:56.4357584Z   "digest": "sha256:3197820c25f93113f22a6d90d6dbcf70e1d71ae528c3c0b1542e9604bdfa9d83",
2025-04-27T10:41:56.4358658Z   "architectures": [
2025-04-27T10:41:56.4359147Z     "linux/amd64",
2025-04-27T10:41:56.4358912Z     "linux/arm64"
2025-04-27T10:41:56.4359362Z   ]
2025-04-27T10:41:56.4359560Z }
2025-04-27T10:41:56.4359897Z ---END OF IMAGE BUILD REPORT---
2025-04-27T10:41:56.4361039Z 
2025-04-27T10:41:56.4431442Z ##[section]Finishing: print_image_build_report`
		expectedReport := &BuildReport{
			Status:        "Succeeded",
			IsPushed:      true,
			IsSigned:      false,
			Images:        []string{"europe-docker.pkg.dev/kyma-project/dev/cors-proxy:PR-12975"},
			Digest:        "sha256:3197820c25f93113f22a6d90d6dbcf70e1d71ae528c3c0b1542e9604bdfa9d83",
			Name:          "cors-proxy",
			Tags:          []string{"PR-12975"},
			RegistryURL:   "europe-docker.pkg.dev/kyma-project/dev",
			Architectures: []string{"linux/amd64", "linux/arm64"},
		}

		It("parses the image build report", func() {
			actual, err := NewBuildReportFromLogs(logs)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual).To(Equal(expectedReport))
		})

		It("correctly parses digest as a string", func() {
			actual, err := NewBuildReportFromLogs(logs)
			Expect(err).ToNot(HaveOccurred())

			Expect(actual.Digest).To(BeAssignableToTypeOf(""))
			Expect(actual.Digest).To(Equal("sha256:3197820c25f93113f22a6d90d6dbcf70e1d71ae528c3c0b1542e9604bdfa9d83"))
			Expect(actual.Digest).To(HavePrefix("sha256:"))
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
