package bootstrap

import "github.com/alexperezortuno/portx/internal/cli"

type Bootstrap struct {
	root *cli.RootCommand
}

func New() (*Bootstrap, error) {

	root := cli.NewRoot()

	return &Bootstrap{
		root: root,
	}, nil
}

func (a *Bootstrap) Run(args []string) error {
	return a.root.Execute(args)
}
