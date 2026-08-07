// Package output renders results as table, JSON, or CSV.
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Encoder writes values in the requested format.
type Encoder struct {
	Format string
	Out    io.Writer
}

// New returns an Encoder writing to stdout.
func New(format string) *Encoder {
	if format == "" {
		format = "table"
	}
	return &Encoder{Format: format, Out: os.Stdout}
}

// WriteValue encodes a single value as JSON (default for complex) or table/csv.
func (e *Encoder) WriteValue(v any) error {
	switch strings.ToLower(e.Format) {
	case "json":
		enc := json.NewEncoder(e.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	case "csv":
		return e.writeCSV(v)
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(e.Out, string(b))
		return err
	}
}

func (e *Encoder) writeCSV(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var rows []map[string]any
	if err := json.Unmarshal(b, &rows); err != nil {
		var one map[string]any
		if err2 := json.Unmarshal(b, &one); err2 != nil {
			return fmt.Errorf("csv output requires object or array of objects")
		}
		rows = []map[string]any{one}
	}
	if len(rows) == 0 {
		return nil
	}
	headers := make([]string, 0)
	seen := map[string]bool{}
	for _, r := range rows {
		for k := range r {
			if !seen[k] {
				seen[k] = true
				headers = append(headers, k)
			}
		}
	}
	w := csv.NewWriter(e.Out)
	if err := w.Write(headers); err != nil {
		return err
	}
	for _, r := range rows {
		rec := make([]string, len(headers))
		for i, h := range headers {
			rec[i] = fmt.Sprint(r[h])
		}
		if err := w.Write(rec); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// WriteRows writes tabular data with explicit headers.
func (e *Encoder) WriteRows(headers []string, rows [][]string) error {
	switch strings.ToLower(e.Format) {
	case "json":
		out := make([]map[string]string, 0, len(rows))
		for _, r := range rows {
			m := map[string]string{}
			for i, h := range headers {
				if i < len(r) {
					m[h] = r[i]
				}
			}
			out = append(out, m)
		}
		enc := json.NewEncoder(e.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	case "csv":
		w := csv.NewWriter(e.Out)
		if err := w.Write(headers); err != nil {
			return err
		}
		for _, r := range rows {
			if err := w.Write(r); err != nil {
				return err
			}
		}
		w.Flush()
		return w.Error()
	default:
		tw := tabwriter.NewWriter(e.Out, 0, 4, 2, ' ', 0)
		fmt.Fprintln(tw, strings.Join(headers, "\t"))
		for _, r := range rows {
			fmt.Fprintln(tw, strings.Join(r, "\t"))
		}
		return tw.Flush()
	}
}

// Exit codes used across the CLI.
const (
	ExitOK         = 0
	ExitError      = 1
	ExitUsage      = 2
	ExitDependency = 3
	ExitNotFound   = 4
)
