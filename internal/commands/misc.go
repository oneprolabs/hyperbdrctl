package commands

import "fmt"

func errUnknown(group, cmd string) error {
	if cmd == "" {
		return fmt.Errorf("%s requires subcommand", group)
	}
	return fmt.Errorf("unknown %s command %q", group, cmd)
}

func missing(ctx *context, key string) error {
	return fmt.Errorf(ctx.loc.T(key))
}
