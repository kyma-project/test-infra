package testinfra_test

import (
	"github.com/kyma-project/test-infra/development/tools/jobs/tester/preset"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/jobs/tester"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dnsWebhookJobPath = "./../../../../../prow/jobs/incubator/kyma-dns-webhook/kyma-dns-webhook.yaml"

func TestDNSWebhookJobsPresubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(dnsWebhookJobPath)

	// then
	require.NoError(t, err)

	// Webhook
	webhookPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-incubator/kyma-dns-webhook"], "pre-master-kyma-incubator-dns-webhook", "master")
	require.NotNil(t, webhookPresubmit)

	assert.Equal(t, 10, webhookPresubmit.MaxConcurrency)
	assert.False(t, webhookPresubmit.SkipReport)
	assert.True(t, webhookPresubmit.Decorate)
	assert.Equal(t, "^dns-webhook/", webhookPresubmit.RunIfChanged)
	assert.Equal(t, "github.com/kyma-incubator/kyma-dns-webhook", webhookPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, webhookPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, webhookPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, webhookPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/kyma-dns-webhook/dns-webhook"}, webhookPresubmit.Spec.Containers[0].Args)

	// Challenger
	challengerPresubmit := tester.FindPresubmitJobByNameAndBranch(jobConfig.Presubmits["kyma-incubator/kyma-dns-webhook"], "pre-master-kyma-incubator-dns-challenger", "master")
	require.NotNil(t, challengerPresubmit)

	assert.Equal(t, 10, challengerPresubmit.MaxConcurrency)
	assert.False(t, challengerPresubmit.SkipReport)
	assert.True(t, challengerPresubmit.Decorate)
	assert.Equal(t, "^dns-challenger/", challengerPresubmit.RunIfChanged)
	assert.Equal(t, "github.com/kyma-incubator/kyma-dns-webhook", challengerPresubmit.PathAlias)
	tester.AssertThatHasPresets(t, challengerPresubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush, preset.BuildPr)
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, challengerPresubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, challengerPresubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/kyma-dns-webhook/dns-challenger"}, challengerPresubmit.Spec.Containers[0].Args)
}

func TestDNSWebhookJobsPostsubmit(t *testing.T) {
	// when
	jobConfig, err := tester.ReadJobConfig(dnsWebhookJobPath)

	// then
	require.NoError(t, err)

	// Webhook
	webhookPostsubmit := tester.FindPostsubmitJobByNameAndBranch(jobConfig.Postsubmits["kyma-incubator/kyma-dns-webhook"], "post-master-kyma-incubator-dns-webhook", "master")
	require.NotNil(t, webhookPostsubmit)

	assert.Equal(t, 10, webhookPostsubmit.MaxConcurrency)
	assert.True(t, webhookPostsubmit.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/kyma-dns-webhook", webhookPostsubmit.PathAlias)
	tester.AssertThatHasPresets(t, webhookPostsubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush, preset.BuildMaster)
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, webhookPostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, webhookPostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/kyma-dns-webhook/dns-webhook"}, webhookPostsubmit.Spec.Containers[0].Args)

	// Challenger
	challengerPostsubmit := tester.FindPostsubmitJobByNameAndBranch(jobConfig.Postsubmits["kyma-incubator/kyma-dns-webhook"], "post-master-kyma-incubator-dns-challenger", "master")
	require.NotNil(t, challengerPostsubmit)

	assert.Equal(t, 10, challengerPostsubmit.MaxConcurrency)
	assert.True(t, challengerPostsubmit.Decorate)
	assert.Equal(t, "github.com/kyma-incubator/kyma-dns-webhook", challengerPostsubmit.PathAlias)
	tester.AssertThatHasPresets(t, challengerPostsubmit.JobBase, preset.DindEnabled, preset.DockerPushRepoIncubator, preset.GcrPush, preset.BuildMaster)
	assert.Equal(t, tester.ImageGolangKubebuilder2BuildpackLatest, challengerPostsubmit.Spec.Containers[0].Image)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-project/test-infra/prow/scripts/build.sh"}, challengerPostsubmit.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/home/prow/go/src/github.com/kyma-incubator/kyma-dns-webhook/dns-challenger"}, challengerPostsubmit.Spec.Containers[0].Args)
}
