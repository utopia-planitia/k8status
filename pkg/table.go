package k8status

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
