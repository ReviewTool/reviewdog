package client

import (
	"context"

	"github.com/google/go-github/v53/github"

	"github.com/reviewtool/reviewdog/doghouse"
	"github.com/reviewtool/reviewdog/doghouse/server"
)

// GitHubClient is client which talks to GitHub directly instead of talking to
// doghouse server.
type GitHubClient struct {
	Client *github.Client
}

func (c *GitHubClient) Check(ctx context.Context, req *doghouse.CheckRequest) (*doghouse.CheckResponse, error) {
	return server.NewChecker(req, c.Client).Check(ctx)
}
