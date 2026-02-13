package imagebuilder

import (
	"encoding/json"

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

	Describe("BuildReport JSON marshaling", func() {
		report := &BuildReport{
			Status:        "Succeeded",
			IsPushed:      true,
			IsSigned:      false,
			Name:          "my-image",
			Images:        []string{"europe-docker.pkg.dev/kyma-project/prod/my-image:v20260213-abc12345"},
			Digest:        "sha256:d3e4b9ad13d47bb5ee85804cce30f8f2fca16cbd4c0717b0c5db35299ef8ccef",
			Tags:          []string{"v20260213-abc12345", "latest"},
			RegistryURL:   "europe-docker.pkg.dev/kyma-project/prod",
			Architectures: []string{"linux/arm64", "linux/amd64"},
		}

		It("marshals build report to valid JSON", func() {
			jsonBytes, err := json.Marshal(report)
			Expect(err).ToNot(HaveOccurred())
			Expect(jsonBytes).ToNot(BeEmpty())

			// Verify it's valid JSON by unmarshaling back
			var unmarshaled BuildReport
			err = json.Unmarshal(jsonBytes, &unmarshaled)
			Expect(err).ToNot(HaveOccurred())
			Expect(unmarshaled).To(Equal(*report))
		})

		It("contains all expected fields in JSON output", func() {
			jsonBytes, err := json.Marshal(report)
			Expect(err).ToNot(HaveOccurred())

			var jsonMap map[string]interface{}
			err = json.Unmarshal(jsonBytes, &jsonMap)
			Expect(err).ToNot(HaveOccurred())

			// Verify all fields are present
			Expect(jsonMap).To(HaveKey("status"))
			Expect(jsonMap).To(HaveKey("pushed"))
			Expect(jsonMap).To(HaveKey("signed"))
			Expect(jsonMap).To(HaveKey("image_name"))
			Expect(jsonMap).To(HaveKey("images_list"))
			Expect(jsonMap).To(HaveKey("digest"))
			Expect(jsonMap).To(HaveKey("tags"))
			Expect(jsonMap).To(HaveKey("repository_path"))
			Expect(jsonMap).To(HaveKey("architectures"))

			// Verify field values
			Expect(jsonMap["status"]).To(Equal("Succeeded"))
			Expect(jsonMap["pushed"]).To(BeTrue())
			Expect(jsonMap["signed"]).To(BeFalse())
			Expect(jsonMap["image_name"]).To(Equal("my-image"))
			Expect(jsonMap["digest"]).To(Equal("sha256:d3e4b9ad13d47bb5ee85804cce30f8f2fca16cbd4c0717b0c5db35299ef8ccef"))
			Expect(jsonMap["repository_path"]).To(Equal("europe-docker.pkg.dev/kyma-project/prod"))
		})

		It("marshals arrays correctly", func() {
			jsonBytes, err := json.Marshal(report)
			Expect(err).ToNot(HaveOccurred())

			var jsonMap map[string]interface{}
			err = json.Unmarshal(jsonBytes, &jsonMap)
			Expect(err).ToNot(HaveOccurred())

			// Verify arrays
			images := jsonMap["images_list"].([]interface{})
			Expect(images).To(HaveLen(1))
			Expect(images[0]).To(Equal("europe-docker.pkg.dev/kyma-project/prod/my-image:v20260213-abc12345"))

			tags := jsonMap["tags"].([]interface{})
			Expect(tags).To(HaveLen(2))
			Expect(tags).To(ContainElements("v20260213-abc12345", "latest"))

			architectures := jsonMap["architectures"].([]interface{})
			Expect(architectures).To(HaveLen(2))
			Expect(architectures).To(ContainElements("linux/arm64", "linux/amd64"))
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
