package catalog

import (
	"os"
	"path/filepath"

	"github.com/liaotuo/lt-clean/internal/sysutil"
)

// Build returns the full catalog of cleanable items (31 entries).
//
// Mirrors the Rust build_catalog() in lt-tool/src/cleaner.rs.
func Build() []Item {
	home := sysutil.Home()
	if home == "" {
		home = "/Users/lt"
	}
	projectRoot, err := os.Getwd()
	if err != nil {
		projectRoot = "."
	}

	probePaths := func(item *Item) bool {
		for _, p := range item.SizePaths {
			if _, err := os.Stat(p); err == nil {
				return true
			}
		}
		return false
	}
	probeCmd := func(name string) func(*Item) bool {
		return func(*Item) bool { return sysutil.CommandAvailable(name) }
	}
	probeAlways := func(*Item) bool { return true }

	items := []Item{
		// ── dev_caches ────────────────────────────────────────────────────
		{
			ID: "brew", Group: "dev_caches", Title: "Homebrew", Level: Safe,
			SizePaths: []string{filepath.Join(home, "Library/Caches/Homebrew")},
			Action:    Action{Kind: ActCmd, Program: "brew", Args: []string{"cleanup", "--prune=all"}},
			Probe:     probeCmd("brew"),
		},
		{
			ID: "go_modcache", Group: "dev_caches", Title: "Go 模块", Level: Safe,
			SizePaths: []string{filepath.Join(home, "go/pkg/mod/cache")},
			Action:    Action{Kind: ActCmd, Program: "go", Args: []string{"clean", "-modcache"}},
			Probe:     probeCmd("go"),
		},
		{
			ID: "pip", Group: "dev_caches", Title: "pip", Level: Safe,
			SizePaths: []string{filepath.Join(home, "Library/Caches/pip")},
			Action: func() Action {
				if sysutil.CommandAvailable("pip3") {
					return Action{Kind: ActCmd, Program: "pip3", Args: []string{"cache", "purge"}}
				}
				return Action{Kind: ActCmd, Program: "pip", Args: []string{"cache", "purge"}}
			}(),
			Probe: func(*Item) bool {
				return sysutil.CommandAvailable("pip3") || sysutil.CommandAvailable("pip")
			},
		},
		{
			ID: "conda", Group: "dev_caches", Title: "Conda", Level: Safe,
			SizePaths: []string{
				filepath.Join(home, "miniconda3"),
				filepath.Join(home, "anaconda3"),
			},
			Action: Action{Kind: ActCmd, Program: "conda", Args: []string{"clean", "--all", "-y"}},
			Probe:  probeCmd("conda"),
		},
		{
			ID: "npm", Group: "dev_caches", Title: "npm", Level: Safe,
			SizePaths: []string{filepath.Join(home, "Library/Caches/npm")},
			Action:    Action{Kind: ActCmd, Program: "npm", Args: []string{"cache", "clean", "--force"}},
			Probe:     probeCmd("npm"),
		},
		{
			ID: "bun", Group: "dev_caches", Title: "Bun", Level: Safe,
			SizePaths: []string{
				filepath.Join(home, ".bun/install/cache"),
				filepath.Join(home, "Library/Caches/bun"),
			},
			Action: Action{
				Kind: ActMultiPath,
				Paths: []string{
					filepath.Join(home, ".bun/install/cache"),
					filepath.Join(home, "Library/Caches/bun"),
				},
				Program: "bun",
				Args:    []string{"pm", "cache", "rm"},
			},
			Probe: probeCmd("bun"),
		},
		{
			ID: "yarn", Group: "dev_caches", Title: "Yarn", Level: Safe,
			SizePaths: []string{filepath.Join(home, "Library/Caches/Yarn")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/Yarn")}},
			Probe:     probePaths,
		},
		{
			ID: "pnpm", Group: "dev_caches", Title: "pnpm", Level: Safe,
			SizePaths: nil,
			Action:    Action{Kind: ActCmd, Program: "pnpm", Args: []string{"store", "prune"}},
			Probe:     probeCmd("pnpm"),
		},
		{
			ID: "cargo_registry", Group: "dev_caches", Title: "Cargo registry", Level: Safe,
			SizePaths: []string{
				filepath.Join(home, ".cargo/registry/cache"),
				filepath.Join(home, ".cargo/registry/src"),
			},
			Action: Action{
				Kind: ActMultiPath,
				Paths: []string{
					filepath.Join(home, ".cargo/registry/cache"),
					filepath.Join(home, ".cargo/registry/src"),
				},
			},
			Probe: probePaths,
		},
		{
			ID: "node_gyp", Group: "dev_caches", Title: "node-gyp", Level: Safe,
			SizePaths: []string{filepath.Join(home, "Library/Caches/node-gyp")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/node-gyp")}},
			Probe:     probePaths,
		},
		{
			ID: "typescript", Group: "dev_caches", Title: "TypeScript", Level: Safe,
			SizePaths: []string{filepath.Join(home, "Library/Caches/typescript")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/typescript")}},
			Probe:     probePaths,
		},
		{
			ID: "playwright", Group: "dev_caches", Title: "Playwright 浏览器", Level: Costly,
			SizePaths: []string{filepath.Join(home, "Library/Caches/ms-playwright")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/ms-playwright")}},
			Probe:     probePaths,
		},
		{
			ID: "cypress", Group: "dev_caches", Title: "Cypress", Level: Costly,
			SizePaths: []string{filepath.Join(home, "Library/Caches/Cypress")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/Cypress")}},
			Probe:     probePaths,
		},
		{
			ID: "gradle", Group: "dev_caches", Title: "Gradle", Level: Costly,
			SizePaths: []string{filepath.Join(home, ".gradle/caches")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".gradle/caches")}},
			Probe:     probePaths,
		},
		{
			ID: "maven", Group: "dev_caches", Title: "Maven", Level: Costly,
			SizePaths: []string{filepath.Join(home, ".m2/repository")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".m2/repository")}},
			Probe:     probePaths,
		},
		{
			ID: "cocoapods", Group: "dev_caches", Title: "CocoaPods 仓库", Level: Costly,
			SizePaths: []string{filepath.Join(home, ".cocoapods/repos")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".cocoapods/repos")}},
			Probe:     probePaths,
		},
		{
			ID: "xdg_cache", Group: "dev_caches", Title: "XDG 缓存", Level: Costly,
			SizePaths: []string{filepath.Join(home, ".cache")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".cache")}},
			Probe:     probePaths,
		},

		// ── ide ────────────────────────────────────────────────────────────
		{
			ID: "xcode_derived", Group: "ide", Title: "Xcode DerivedData", Level: Safe,
			SizePaths: []string{filepath.Join(home, "Library/Developer/Xcode/DerivedData")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Developer/Xcode/DerivedData")}},
			Probe:     probePaths,
		},
		{
			ID: "vscode_cache", Group: "ide", Title: "VSCode Cache", Level: Safe,
			SizePaths: []string{
				filepath.Join(home, "Library/Application Support/Code/Cache"),
				filepath.Join(home, "Library/Application Support/Code/CachedData"),
				filepath.Join(home, "Library/Application Support/Code/CachedExtensions"),
				filepath.Join(home, "Library/Application Support/Code/logs"),
			},
			Action: Action{
				Kind: ActMultiPath,
				Paths: []string{
					filepath.Join(home, "Library/Application Support/Code/Cache"),
					filepath.Join(home, "Library/Application Support/Code/CachedData"),
					filepath.Join(home, "Library/Application Support/Code/CachedExtensions"),
					filepath.Join(home, "Library/Application Support/Code/logs"),
				},
			},
			Probe: probePaths,
		},
		{
			ID: "jetbrains_cache", Group: "ide", Title: "JetBrains", Level: Safe,
			SizePaths: []string{filepath.Join(home, "Library/Caches/JetBrains")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/JetBrains")}},
			Probe:     probePaths,
		},

		// ── mobile ────────────────────────────────────────────────────────
		{
			ID: "ios_simulator_unavail", Group: "mobile", Title: "iOS Simulator 失效设备", Level: Safe,
			SizePaths: nil,
			Action:    Action{Kind: ActCmd, Program: "xcrun", Args: []string{"simctl", "delete", "unavailable"}},
			Probe:     probeCmd("xcrun"),
		},
		{
			ID: "ios_backup", Group: "mobile", Title: "iOS 设备备份", Level: Destructive,
			SizePaths: []string{filepath.Join(home, "Library/Application Support/MobileSync/Backup")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Application Support/MobileSync/Backup")}},
			Probe:     probePaths,
		},

		// ── system ─────────────────────────────────────────────────────────
		{
			ID: "apfs_snapshots", Group: "system", Title: "APFS 本地快照", Level: Costly,
			SizePaths: nil,
			Action:    Action{Kind: ActCmd, Program: "tmutil", Args: []string{"thinlocalsnapshots", "/", "999999999", "1"}},
			Probe:     probeCmd("tmutil"),
		},
		{
			ID: "user_logs", Group: "system", Title: "用户日志", Level: Safe,
			SizePaths: []string{filepath.Join(home, "Library/Logs")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Logs")}},
			Probe:     probePaths,
		},
		{
			ID: "system_logs_archived", Group: "system", Title: "已归档系统日志", Level: Safe,
			SizePaths: []string{"/private/var/log"},
			Action: Action{
				Kind: ActRmGlobInDir,
				Dir:  "/private/var/log",
				Exts: []string{"gz", "bz2"},
			},
			Probe: probePaths,
		},
		{
			ID: "quicklook_cache", Group: "system", Title: "QuickLook 缩略图", Level: Safe,
			SizePaths: nil,
			Action:    Action{Kind: ActCmd, Program: "qlmanage", Args: []string{"-r", "cache"}},
			Probe:     probeCmd("qlmanage"),
		},
		{
			ID: "trash", Group: "system", Title: "回收站", Level: Destructive,
			SizePaths: []string{filepath.Join(home, ".Trash")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".Trash")}},
			Probe:     probePaths,
		},
		{
			ID: "ds_store", Group: "system", Title: ".DS_Store 递归扫除", Level: Safe,
			SizePaths: []string{home},
			Action:    Action{Kind: ActDsStoreSweep, Paths: []string{home}},
			Probe:     probeAlways,
		},

		// ── project ───────────────────────────────────────────────────────
		{
			ID: "lt_target", Group: "project", Title: "LT Tool target/", Level: Costly,
			SizePaths: []string{filepath.Join(projectRoot, "target")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(projectRoot, "target")}},
			Probe:     probePaths,
		},
		{
			ID: "lt_tmp", Group: "project", Title: "LT Tool tmp/", Level: Safe,
			SizePaths: []string{filepath.Join(projectRoot, "tmp")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(projectRoot, "tmp")}},
			Probe:     probePaths,
		},
		{
			ID: "lt_logs", Group: "project", Title: "LT Tool logs/", Level: Safe,
			SizePaths: []string{filepath.Join(projectRoot, "logs")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(projectRoot, "logs")}},
			Probe:     probePaths,
		},
	}

	return items
}

// FindByID returns the item with the given ID, or nil.
func FindByID(items []Item, id string) *Item {
	for i := range items {
		if items[i].ID == id {
			return &items[i]
		}
	}
	return nil
}
