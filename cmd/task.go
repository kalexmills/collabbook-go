// Copyright © 2018 K. Alex Mills <k.alex.mills@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/kalexmills/collabbook-go/data"
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"t"},
	Short:   "Create task",
	DisableFlagsInUseLine: true,
	Long: `
Creates a new task, optionally assigning it to boards. Any argument starting
with either '@' or '#' are interpreted as the names of boards to which the
task will be added. Boards which do not already exist are created.

Examples:

   cb task A new task
   cb task #1 #2 A task which is on boards 1 and 2
   cb task "#Name has spaces" A task on a board named "Name has spaces"  
`,
	Args: cobra.MinimumNArgs(1),
	Run: ItemCreate("task", func (desc string, boards ...string) (*data.Item) {
		return itemstore.MakeTask(desc, boards...)
	}),
}

func init() {
	rootCmd.AddCommand(taskCmd)
}
