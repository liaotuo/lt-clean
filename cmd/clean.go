package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/liaotuo/lt-clean/internal/catalog"
	"github.com/liaotuo/lt-clean/internal/cleaner"
	"github.com/liaotuo/lt-clean/internal/config"
	"github.com/spf13/cobra"
)

var (
	cleanIDs       []string
	cleanGroup     string
	cleanSafe      bool
	cleanDryRun    bool
	cleanPermanent bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean selected items (non-interactive)",
	Long: "Run cleaning actions for items selected via --id, --group, or --safe.\n" +
		"At least one selector is required. Use --dry-run to preview without deleting.",
	RunE: func(cmd *cobra.Command, args []string) error {
		items := catalog.Build()
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		items = catalog.MergeCustom(items, cfg.Custom)

		ids, err := resolveIDs(items, cleanIDs, cleanGroup, cleanSafe, cfg)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return errors.New("no items selected; pass --id, --group, or --safe")
		}

		mode := cleaner.ModeTrash
		if cleanPermanent {
			mode = cleaner.ModePermanent
		}

		actionWord := "trashed"
		summaryWord := "trashed"
		if cleanPermanent {
			actionWord = "freed"
			summaryWord = "freed"
		}

		modeTag := " [TRASH]"
		if cleanPermanent {
			modeTag = " [PERMANENT]"
		}

		fmt.Printf("cleaning %d items%s%s:\n", len(ids), dryRunSuffix(cleanDryRun), modeTag)
		summary := cleaner.Run(items, ids, mode, cleanDryRun, func(p cleaner.Progress) {
			switch p.Status {
			case "ok":
				fmt.Printf("  ✓ %-32s  %s %s\n", p.ID, actionWord, humanize.Bytes(uint64(p.FreedBytes)))
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

		fmt.Printf("\ntotal %s: %s   success: %d   failed: %d\n",
			summaryWord,
			humanize.Bytes(uint64(summary.TotalFreed)),
			summary.SuccessCount, summary.FailCount)
		if summary.FailCount > 0 {
			return fmt.Errorf("%d items failed", summary.FailCount)
		}
		return nil
	},
}

func resolveIDs(items []catalog.Item, ids []string, group string, safeOnly bool, cfg config.Config) ([]string, error) {
	set := make(map[string]bool)

	// --id is explicit user intent: never filtered by config.
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

	var skipped []string

	if group != "" {
		matched := false
		for _, it := range items {
			if it.Group != group || !it.Available() {
				continue
			}
			if cfg.Excluded(it.ID) {
				skipped = append(skipped, it.ID)
				continue
			}
			set[it.ID] = true
			matched = true
		}
		if !matched {
			return nil, fmt.Errorf("no available items in group: %s", group)
		}
	}

	if safeOnly {
		for _, it := range items {
			if it.Level != catalog.Safe || !it.Available() {
				continue
			}
			if cfg.Excluded(it.ID) {
				skipped = append(skipped, it.ID)
				continue
			}
			set[it.ID] = true
		}
	}

	if len(skipped) > 0 {
		fmt.Printf("excluded by config: %s\n", strings.Join(uniqueStrings(skipped), ", "))
	}

	out := make([]string, 0, len(set))
	for _, it := range items {
		if set[it.ID] {
			out = append(out, it.ID)
		}
	}
	return out, nil
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func dryRunSuffix(dry bool) string {
	if dry {
		return " [DRY RUN]"
	}
	return ""
}

func init() {
	cleanCmd.Flags().StringSliceVar(&cleanIDs, "id", nil, "comma-separated item ids (e.g. brew,go_modcache)")
	cleanCmd.Flags().StringVar(&cleanGroup, "group", "", "select all items in a group (dev_caches|ide|mobile|system)")
	cleanCmd.Flags().BoolVar(&cleanSafe, "safe", false, "select all Safe-level items")
	cleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "preview without deleting")
	cleanCmd.Flags().BoolVar(&cleanPermanent, "permanent", false, "permanently delete instead of moving to ~/.Trash")
	rootCmd.AddCommand(cleanCmd)
}
