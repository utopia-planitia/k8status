package k8status

import "github.com/olekukonko/tablewriter"

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

func tableHeader(t tableAble) []string {
	return t.header()
}

func tableRow(t tableAble) []string {
	return t.row()
}

func CreateTable(writer colorWriter, headers []string, titleColor int) (*tablewriter.Table, error) {

	table := tablewriter.NewWriter(writer)
	table.SetHeader(headers)
	if !writer.noColors {
		headerColors := []tablewriter.Colors{}
		for i := 0; i < len(headers); i++ {
			headerColors = append(headerColors, tablewriter.Colors{tablewriter.Bold, titleColor})
		}
		table.SetHeaderColor(headerColors...)
	}
	table.SetHeaderLine(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	_, err := writer.Write([]byte("\n"))
	if err != nil {
		return table, err
	}

	return table, err
}

func RenderTable(table *tablewriter.Table, data [][]string) {
	table.AppendBulk(data)
	table.Render()
}
