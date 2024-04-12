package main

import (
	"github.com/kyma-project/test-infra/pkg/prow/externalplugin"
	"sigs.k8s.io/prow/prow/git/v2"
)

const (
	PluginName = "needs-tws"
)

func main() {
	l := externalplugin.NewLogger()
	o := externalplugin.Opts{}
	fs := o.NewFlags()
	o.ParseFlags(fs)

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
	p.RegisterWebhookHandler("pull_request", pb.PullRequestHandler)
	p.RegisterWebhookHandler("pull_request_review", pb.PullRequestReviewHandler)

	externalplugin.Start(&p, HelpProvider, &o)
}
