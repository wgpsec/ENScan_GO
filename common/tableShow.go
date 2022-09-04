package common

import (
	"github.com/olekukonko/tablewriter"
	"os"
)

func TableShow(keys []string, values [][]string, options *ENOptions) {
	if !options.IsApiMode {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAlignment(tablewriter.ALIGN_CENTER)
		table.SetHeader(keys)
		table.AppendBulk(values)
		table.Render()
	}
}
