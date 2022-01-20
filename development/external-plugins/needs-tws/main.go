package main

import (
	"github.com/kyma-project/test-infra/development/prow/externalplugin"
	"k8s.io/test-infra/prow/git/v2"
)

const (
	PluginName = "needs-tws"
)

func main() {
	l := externalplugin.NewLogger()
	o := externalplugin.Opts{}
	fs := o.GatherDefaultOptions()
	o.Parse(fs)

	ghc, err := o.Github.GitHubClient(o.DryRun)
	if err != nil {
		l.Fatalw("Could not get github client.", "error", err)
	}
	g, err := o.Github.GitClient(o.DryRun)
	if err != nil {
		l.Fatalw("Could not get git client.", "error", err)
	}
	gc := git.ClientFactoryFrom(g)
	pb := PluginBackend{
		ghc: ghc,
		gcf: gc,
		oac: AliasesClient{},
	}

	p := externalplugin.Plugin{
		Name: PluginName,
	}
	p.WithLogger(l)
	p.WithWebhookSecret(o.WebhookSecretPath)
	p.RegisterWebhookHandler("pull_request", pb.HandlePR)
	p.RegisterWebhookHandler("pull_request_review", pb.HandlePRReview)

	externalplugin.Start(&p, HelpProvider, &o)
}
