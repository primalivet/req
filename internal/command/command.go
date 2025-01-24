package command

import (
	"flag"
)

type Name string

const (
	HTTP Name = "http"
	GraphQL Name = "gql"
)

type Handler = func(fs *flag.FlagSet, args []string) error

type Command struct {
	name Name
	fs *flag.FlagSet
	handler Handler 
}

func New(name Name, fs *flag.FlagSet, handle Handler) *Command {
	c := &Command{
		name: name,
		fs: flag.NewFlagSet(string(name), flag.ExitOnError),
		handler: handle,
	}
	return c
}

func (c *Command) Run(args []string) error {
	err := c.handler(c.fs, args)
	return err
}

