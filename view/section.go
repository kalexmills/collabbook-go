package view

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/kalexmills/collabbook-go/data"
	"strings"
	"strconv"
	"os"
)

type Section struct {
	Heading *string
	Items   []uint64
}

var done, notes, tasks int

var sections []Section

func PrintSections(store *data.Repo, factory func() []Section) {
	sections = factory()

	if len(sections) == 0 {
		hoorayNothingToDo()
	}

	printSections(store)
	printFooter()
}

func printSections(store *data.Repo) {
	allDone := true
	for _, section := range sections {
		if len(section.Items) > 0 {
			allDone = false
			sDone, sNotes, sTasks := countAll(store, section.Items)
			done += sDone
			notes += sNotes
			tasks += sTasks

			printBoardHeading(*section.Heading, sDone, sTasks)
			for _, id := range section.Items {
				if item := store.Item(id); item != nil {
					printItem(item)
				}
			}
			fmt.Println()
		}
	}
	if allDone {
		hoorayNothingToDo()
	}
}

func hoorayNothingToDo() {
	Success(`\(^_^)/`, "All done!")
	os.Exit(0)
}

func countAll(store *data.Repo, items []uint64) (done, notes, tasks int) {
	for _, id := range items {
		item := store.Item(id)
		if item == nil {
			continue
		}
		if item.IsTask() {
			tasks += 1
			if item.IsComplete() {
				done += 1
			}
		} else {
			notes += 1
		}
	}
	return
}

func printItem(it *data.Item) {
	star := star(it)
	fmt.Fprintf(color.Output, "  %4d. %s %s %s %s\n", it.Id, checkbox(it), star, description(it), star)
}

func printFooter() {
	pending := tasks - done
	var pct int
	if tasks == 0 {
		pct = 100
	} else {
		pct = (int)((100.0*done)/(1.0*tasks))
	}

	fmt.Fprintf(color.Output, "  %d%% of all tasks complete.\n", pct)
	fmt.Fprintln(color.Output, "  "+strings.Join([]string{
		Green(strconv.Itoa(done)) + " done",
		Yellow(strconv.Itoa(pending)) + " pending",
		Blue(strconv.Itoa(notes)) + " notes",
	}, " - "))
}

func printBoardHeading(name string, complete int, total int) {
	fmt.Fprintf(color.Output, "  %s [%d/%d]\n", White(name), complete, total)
}

func star(it *data.Item) string {
	if it.IsStarred() {
		return Yellow("** ")
	}
	return ""
}

func checkbox(it *data.Item) string {
	if !it.IsTask() {
		return " - "
	}
	if it.IsComplete() {
		return "[" + Green("X") + "]"
	}
	return "[ ]"
}

func description(it *data.Item) string {
	if it.IsTask() && it.IsComplete() {
		return it.Desc
	}
	if it.IsStarred() {
		return Yellow(it.Desc)
	}
	return it.Desc
}
