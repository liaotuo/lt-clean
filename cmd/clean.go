package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/liaotuo/lt-clean/internal/catalog"
	"github.com/liaotuo/lt-clean/internal/cleaner"
	"github.com/spf13/cobra"
)

var (
	cleanIDs    []string
	cleanGroup  string
	cleanSafe   bool
	cleanDryRun bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean selected items (non-interactive)",
	Long: "Run cleaning actions for items selected via --id, --group, or --safe.\n" +
		"At least one selector is required. Use --dry-run to preview without deleting.",
	RunE: func(cmd *cobra.Command, args []string) error {
		items := catalog.Build()

		ids, err := resolveIDs(items, cleanIDs, cleanGroup, cleanSafe)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return errors.New("no items selected; pass --id, --group, or --safe")
		}

		fmt.Printf("cleaning %d items%s:\n", len(ids), dryRunSuffix(cleanDryRun))
		summary := cleaner.Run(items, ids, cleanDryRun, func(p cleaner.Progress) {
			switch p.Status {
			case "ok":
				fmt.Printf("  ✓ %-32s  freed %s\n", p.ID, humanize.Bytes(uint64(p.FreedBytes)))
			case "fail":
				msg := ""
				if p.Err != nil {
					msg = p.Err.Error()
				}
				fmt.Printf("  ✗ %-32s  %s\n", p.ID, msg)
			case "dryrun":
				fmt.Printf("  ○ %-32s  (dry-run)\n", p.ID)
			}
		})

		fmt.Printf("\ntotal freed: %s   success: %d   failed: %d\n",
			humanize.Bytes(uint64(summary.TotalFreed)),
			summary.SuccessCount, summary.FailCount)
		if summary.FailCount > 0 {
			return fmt.Errorf("%d items failed", summary.FailCount)
		}
		return nil
	},
}

func resolveIDs(items []catalog.Item, ids []string, group string, safeOnly bool) ([]string, error) {
	set := make(map[string]bool)

	for _, raw := range ids {
		for _, id := range strings.Split(raw, ",") {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if catalog.FindByID(items, id) == nil {
				return nil, fmt.Errorf("unknown id: %s", id)
			}
			set[id] = true
		}
	}

	if group != "" {
		matched := false
		for _, it := range items {
			if it.Group == group && it.Available() {
				set[it.ID] = true
				matched = true
			}
		}
		if !matched {
			return nil, fmt.Errorf("no available items in group: %s", group)
		}
	}

	if safeOnly {
		for _, it := range items {
			if it.Level == catalog.Safe && it.Available() {
				set[it.ID] = true
			}
		}
	}

	out := make([]string, 0, len(set))
	for _, it := range items {
		if set[it.ID] {
			out = append(out, it.ID)
		}
	}
	return out, nil
}

func dryRunSuffix(dry bool) string {
	if dry {
		return " [DRY RUN]"
	}
	return ""
}

func init() {
	cleanCmd.Flags().StringSliceVar(&cleanIDs, "id", nil, "comma-separated item ids (e.g. brew,go_modcache)")
	cleanCmd.Flags().StringVar(&cleanGroup, "group", "", "select all items in a group (dev_caches|ide|mobile|system|project)")
	cleanCmd.Flags().BoolVar(&cleanSafe, "safe", false, "select all Safe-level items")
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "preview without deleting")
	rootCmd.AddCommand(cleanCmd)
}
