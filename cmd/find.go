package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/liaotuo/lt-clean/internal/finder"
	"github.com/liaotuo/lt-clean/internal/findtui"
	"github.com/liaotuo/lt-clean/internal/sysutil"
	"github.com/spf13/cobra"
)

var (
	findMinSize string
	findJSON    bool
	findTopN    int
)

var findCmd = &cobra.Command{
	Use:   "find [path]",
	Short: "查找大文件并分析磁盘空间分布",
	Long: `查找指定目录下的大文件并分析磁盘空间分布。
默认路径为用户主目录 (~)。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("accepts at most 1 arg(s), received %d", len(args))
		}

		path := sysutil.Home()
		if len(args) == 1 {
			path = sysutil.ExpandHome(args[0])
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("path not found: %s", path)
		}

		var minSize int64
		if findMinSize != "" {
			var err error
			minSize, err = sysutil.ParseSizeSuffix(findMinSize)
			if err != nil {
				return err
			}
		}

		if findJSON || findTopN > 0 {
			cfg := finder.Config{
				Root:    path,
				MinSize: minSize,
				TopN:    findTopN,
			}
			ch, err := finder.Walk(context.Background(), cfg)
			if err != nil {
				return err
			}

			var result finder.WalkResult
			for r := range ch {
				result = r
			}
			if result.Err != nil {
				return result.Err
			}

			var topFiles []finder.FileEntry
			if findTopN > 0 {
				topFiles = finder.TopNFiles(result.Tree, findTopN)
			}

			output, err := finder.MarshalJSON(topFiles)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		}

		cfg := finder.Config{
			Root:    path,
			MinSize: minSize,
		}
		p := tea.NewProgram(findtui.New(cfg))
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	findCmd.Flags().StringVar(&findMinSize, "min-size", "", "minimum file size (e.g. 10M, 1G)")
	findCmd.Flags().BoolVar(&findJSON, "json", false, "emit machine-readable JSON")
	findCmd.Flags().IntVar(&findTopN, "top", 0, "show top N largest files")
	rootCmd.AddCommand(findCmd)
}
