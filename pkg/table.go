package k8status

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

type Table struct {
	Header []string
	Rows   [][]string
}

func (t Table) Fprint(w io.Writer, colored bool) error {
	if len(t.Rows) == 0 {
		return nil
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader(t.Header)

	if colored {
		titleColor := tablewriter.Colors{tablewriter.Bold, tablewriter.FgYellowColor}
		headerColors := []tablewriter.Colors{}
		for i := 0; i < len(t.Header); i++ {
			headerColors = append(headerColors, titleColor)
		}
		table.SetHeaderColor(headerColors...)
	}

	table.SetHeaderLine(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	_, err := w.Write([]byte("\n"))
	if err != nil {
		return err
	}

	table.AppendBulk(t.Rows)
	table.Render()

	return nil
}
