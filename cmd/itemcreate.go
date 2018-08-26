package cmd

import (
	"github.com/kalexmills/collabbook-go/data"
	"strings"
	"github.com/spf13/cobra"
	"os"
	"github.com/kalexmills/collabbook-go/view"
	"strconv"
)

func ItemCreate(name string, factory func(desc string, boards ...string) (*data.Item)) func (*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		var b strings.Builder
		boards := make([]string, 0)
		// Extract board names from arguments
		for _, arg := range args {
			if arg[0] == '#' || arg[0] == '@' {
				boards = append(boards[1:], arg)
			} else {
				b.WriteString(arg)
				b.WriteRune(' ')
			}
		}
		desc := b.String()
		if len(desc) == 0 {
			view.Failure(`:-\`, "No description found for your " + name)
			os.Exit(1)
		}

		item := factory(desc,boards...)
		view.Success(`:-)`, "Created " + name + ": " + strconv.FormatUint(item.Id, 10))
	}
}