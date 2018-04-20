package prompt

import (
	"context"

	"github.com/peco/peco"
)

type pecoCollectResults interface {
	CollectResults() bool
}

// WithPeco picks using Peco
func WithPeco(message string, list []string, defaultOpt string) (string, error) {
	// TODO: Use message
	// TODO: use List
	// TODO: set default
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli := peco.New()
	cli.Argv = []string{}

	err := cli.Run(ctx)
	if err != nil {
		if _, ok := err.(pecoCollectResults); !ok {
			return "", err
		}
	}

	ln := cli.Location().LineNumber()
	l, err := cli.CurrentLineBuffer().LineAt(ln)
	if err != nil {
		return "", err
	}

	o := l.Output()
	return o, nil
}
