package k8status

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

type tableAble interface {
	// head should return the header column for a struct (which implements tableAble)
	// e.g. `return []string{"date", "visits"}``
	header() []string
	// row normally should creates a []string using the structs fields
	// (which implements tableAble), representing a table row
	//
	// example:
	//
	// type TableView struct {
	//   date: time.Time
	//   visits: int
	// }
	//
	// func (t TableView) row() {
	// 	return []string{t.date.String(), strconv.Itoa(t.visits)}
	// }
	row() []string
}

type Table struct {
	Header []string
	Rows   [][]string
}

func (t Table) Fprint(w io.Writer, colored bool) error {
	if len(t.Rows) != 0 {
		return nil
	}

	table, err := CreateTable(w, t.Header, colored)
	if err != nil {
		return err
	}

	RenderTable(table, t.Rows)

	return nil
}

func CreateTable(w io.Writer, headers []string, colored bool) (*tablewriter.Table, error) {
	table := tablewriter.NewWriter(w)
	table.SetHeader(headers)

	if colored {
		titleColor := tablewriter.Colors{tablewriter.Bold, tablewriter.FgYellowColor}
		headerColors := []tablewriter.Colors{}
		for i := 0; i < len(headers); i++ {
			headerColors = append(headerColors, titleColor)
		}
		table.SetHeaderColor(headerColors...)
	}

	table.SetHeaderLine(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	_, err := w.Write([]byte("\n"))
	if err != nil {
		return nil, err
	}

	return table, nil
}

// RenderTable "renders" (not really) by writing into the details writer
func RenderTable(table *tablewriter.Table, data [][]string) {
	table.AppendBulk(data)
	table.Render()
}

type Stats interface {
	summary(w io.Writer)
	toTable() Table
}

func printStats(
	stats Stats,
	header io.Writer,
	details io.Writer,
	verbose,
	colored bool,
) error {
	stats.summary(header)

	err := stats.toTable().Fprint(details, colored)
	if err != nil {
		return err
	}

	return nil
}
