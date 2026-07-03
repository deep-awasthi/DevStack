package terminal

import (
	"fmt"
	"io"
	"strings"
)

func Table(w io.Writer, headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}
	writeRow(w, widths, headers)
	sep := make([]string, len(widths))
	for i, width := range widths {
		sep[i] = strings.Repeat("-", width)
	}
	writeRow(w, widths, sep)
	for _, row := range rows {
		writeRow(w, widths, row)
	}
}

func writeRow(w io.Writer, widths []int, row []string) {
	for i, width := range widths {
		cell := ""
		if i < len(row) {
			cell = row[i]
		}
		if i > 0 {
			fmt.Fprint(w, "  ")
		}
		fmt.Fprintf(w, "%-*s", width, cell)
	}
	fmt.Fprintln(w)
}
