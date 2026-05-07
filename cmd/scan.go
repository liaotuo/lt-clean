package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/dustin/go-humanize"
	"github.com/liaotuo/lt-clean/internal/catalog"
	"github.com/liaotuo/lt-clean/internal/scanner"
	"github.com/spf13/cobra"
)

var scanJSON bool

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan all catalog items and report sizes (non-interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		items := catalog.Build()
		available := items[:0:0]
		for _, it := range items {
			if it.Available() {
				available = append(available, it)
			}
		}

		var results []scanner.Result
		ch := scanner.Run(context.Background(), available)
		for r := range ch {
			results = append(results, r)
		}

		sort.SliceStable(results, func(i, j int) bool {
			if results[i].Group != results[j].Group {
				return results[i].Group < results[j].Group
			}
			return results[i].SizeBytes > results[j].SizeBytes
		})

		if scanJSON {
			type out struct {
				ID, Group, Title, Level string
				SizeBytes               int64
				Error                   string `json:",omitempty"`
			}
			var rows []out
			for _, r := range results {
				e := ""
				if r.Err != nil {
					e = r.Err.Error()
				}
				rows = append(rows, out{
					ID: r.ID, Group: r.Group, Title: r.Title,
					Level: r.Level.String(), SizeBytes: r.SizeBytes, Error: e,
				})
			}
			return json.NewEncoder(os.Stdout).Encode(rows)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "GROUP\tID\tLEVEL\tSIZE\tTITLE")
		var total int64
		for _, r := range results {
			size := "—"
			if r.SizeBytes > 0 {
				size = humanize.Bytes(uint64(r.SizeBytes))
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				r.Group, r.ID, r.Level, size, r.Title)
			total += r.SizeBytes
		}
		w.Flush()
		fmt.Printf("\ntotal scanned: %s across %d items\n",
			humanize.Bytes(uint64(total)), len(results))
		return nil
	},
}

func init() {
	scanCmd.Flags().BoolVar(&scanJSON, "json", false, "emit machine-readable JSON")
	rootCmd.AddCommand(scanCmd)
}
