package catalog

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/liaotuo/lt-clean/internal/config"
	"github.com/liaotuo/lt-clean/internal/sysutil"
)

// Build returns the full catalog of cleanable items (36 entries).
//
// Mirrors the Rust build_catalog() in lt-tool/src/cleaner.rs, extended with
// macOS dev-cache entries added after the initial port.
//
// Note: huggingface lives under ~/.cache and overlaps with xdg_cache.
// Both items coexist; running them in series is idempotent (the second pass
// finds the path already gone and no-ops).
func Build() []Item {
	home := sysutil.Home()
	if home == "" {
		home = "/Users/lt"
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
	probeWritableDir := func(dir string) func(*Item) bool {
		const accessWrite = 0x2
		return func(*Item) bool {
			if _, err := os.Stat(dir); err != nil {
				return false
			}
			return syscall.Access(dir, accessWrite) == nil
		}
	}

	items := []Item{
		// ── dev_caches ────────────────────────────────────────────────────
		{
			ID: "brew", Group: "dev_caches", Title: "Homebrew", Level: Safe,
			Hint:      "Homebrew 包管理器下载缓存。下次 brew install 会重新下载。",
			SizePaths: []string{filepath.Join(home, "Library/Caches/Homebrew")},
			Action:    Action{Kind: ActCmd, Program: "brew", Args: []string{"cleanup", "--prune=all"}},
			Probe:     probeCmd("brew"),
		},
		{
			ID: "go_modcache", Group: "dev_caches", Title: "Go 模块", Level: Safe,
			Hint:      "Go 模块下载缓存。下次 go build 会重新下载。",
			SizePaths: []string{filepath.Join(home, "go/pkg/mod/cache")},
			Action:    Action{Kind: ActCmd, Program: "go", Args: []string{"clean", "-modcache"}},
			Probe:     probeCmd("go"),
		},
		{
			ID: "pip", Group: "dev_caches", Title: "pip", Level: Safe,
			Hint:      "pip 包下载缓存。下次 pip install 会重新下载。",
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
			Hint: "Conda 安装目录下的索引和包缓存。重装或恢复需要重新下载（数 GB）。",
			SizePaths: []string{
				filepath.Join(home, "miniconda3"),
				filepath.Join(home, "anaconda3"),
			},
			Action: Action{Kind: ActCmd, Program: "conda", Args: []string{"clean", "--all", "-y"}},
			Probe:  probeCmd("conda"),
		},
		{
			ID: "npm", Group: "dev_caches", Title: "npm", Level: Safe,
			Hint:      "npm 包下载缓存。下次 npm install 会重新下载。",
			SizePaths: []string{filepath.Join(home, "Library/Caches/npm")},
			Action:    Action{Kind: ActCmd, Program: "npm", Args: []string{"cache", "clean", "--force"}},
			Probe:     probeCmd("npm"),
		},
		{
			ID: "pnpm", Group: "dev_caches", Title: "pnpm", Level: Safe,
			Hint:      "pnpm 全局存储的孤儿包。下次安装会按需重新下载。",
			SizePaths: nil,
			Action:    Action{Kind: ActCmd, Program: "pnpm", Args: []string{"store", "prune"}},
			Probe:     probeCmd("pnpm"),
		},
		{
			ID: "js_pkg_caches", Group: "dev_caches", Title: "JS 包管理器缓存", Level: Safe,
			Hint: "Yarn/Bun 包下载缓存目录。下次 install 会重新下载。npm/pnpm 使用独立命令清理。",
			SizePaths: []string{
				filepath.Join(home, "Library/Caches/Yarn"),
				filepath.Join(home, ".bun/install/cache"),
				filepath.Join(home, "Library/Caches/bun"),
			},
			Action: Action{
				Kind: ActMultiPath,
				Paths: []string{
					filepath.Join(home, "Library/Caches/Yarn"),
					filepath.Join(home, ".bun/install/cache"),
					filepath.Join(home, "Library/Caches/bun"),
				},
			},
			Probe: probePaths,
		},
		{
			ID: "cargo_registry", Group: "dev_caches", Title: "Cargo registry", Level: Safe,
			Hint: "Cargo 已下载的源码和压缩包。下次 cargo build 会重新下载（首次较慢）。",
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
			Hint:      "node-gyp 头文件和构建产物缓存。原生模块编译时会重新下载。",
			SizePaths: []string{filepath.Join(home, "Library/Caches/node-gyp")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/node-gyp")}},
			Probe:     probePaths,
		},
		{
			ID: "typescript", Group: "dev_caches", Title: "TypeScript", Level: Safe,
			Hint:      "TypeScript 编译器临时缓存。tsc 会重建。",
			SizePaths: []string{filepath.Join(home, "Library/Caches/typescript")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/typescript")}},
			Probe:     probePaths,
		},
		{
			ID: "playwright", Group: "dev_caches", Title: "Playwright 浏览器", Level: Costly,
			Hint:      "Playwright 自带的浏览器二进制（每个浏览器 ~150 MB）。删除后需 npx playwright install。",
			SizePaths: []string{filepath.Join(home, "Library/Caches/ms-playwright")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/ms-playwright")}},
			Probe:     probePaths,
		},
		{
			ID: "cypress", Group: "dev_caches", Title: "Cypress", Level: Costly,
			Hint:      "Cypress 二进制下载（约 300 MB）。删除后下次启动会重新下载。",
			SizePaths: []string{filepath.Join(home, "Library/Caches/Cypress")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/Cypress")}},
			Probe:     probePaths,
		},
		{
			ID: "gradle", Group: "dev_caches", Title: "Gradle", Level: Costly,
			Hint:      "Gradle 全局缓存（依赖、wrapper）。下次构建会重新下载（数分钟）。",
			SizePaths: []string{filepath.Join(home, ".gradle/caches")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".gradle/caches")}},
			Probe:     probePaths,
		},
		{
			ID: "maven", Group: "dev_caches", Title: "Maven", Level: Costly,
			Hint:      "Maven 本地仓库（.m2/repository）。下次构建会重新下载（数分钟）。",
			SizePaths: []string{filepath.Join(home, ".m2/repository")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".m2/repository")}},
			Probe:     probePaths,
		},
		{
			ID: "cocoapods", Group: "dev_caches", Title: "CocoaPods 仓库", Level: Costly,
			Hint:      "CocoaPods 远程仓库索引。pod install 会重建（首次较慢）。",
			SizePaths: []string{filepath.Join(home, ".cocoapods/repos")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".cocoapods/repos")}},
			Probe:     probePaths,
		},
		{
			ID: "xdg_cache", Group: "dev_caches", Title: "XDG 缓存", Level: Costly,
			Hint:      "~/.cache 通用缓存目录。各类工具按需重建。",
			SizePaths: []string{filepath.Join(home, ".cache")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".cache")}},
			Probe:     probePaths,
		},
		{
			ID: "docker", Group: "dev_caches", Title: "Docker 镜像/构建缓存", Level: Costly,
			Hint:      "Docker 未使用的镜像、容器、网络、构建缓存。常用镜像需要重新拉取。",
			SizePaths: nil,
			Action:    Action{Kind: ActCmd, Program: "docker", Args: []string{"system", "prune", "-a", "-f"}},
			Probe:     probeCmd("docker"),
		},
		{
			ID: "poetry", Group: "dev_caches", Title: "Poetry", Level: Safe,
			Hint:      "Poetry 包下载缓存。下次 poetry install 会重新下载。",
			SizePaths: []string{filepath.Join(home, "Library/Caches/pypoetry")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/pypoetry")}},
			Probe:     probePaths,
		},
		{
			ID: "pyenv", Group: "dev_caches", Title: "pyenv 已装版本", Level: Costly,
			Hint:      "pyenv 已安装的 Python 版本。删除后需 pyenv install 重装（每个版本数分钟）。",
			SizePaths: []string{filepath.Join(home, ".pyenv/versions")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".pyenv/versions")}},
			Probe:     probePaths,
		},
		{
			ID: "pub_cache", Group: "dev_caches", Title: "Dart pub", Level: Safe,
			Hint:      "Dart/Flutter 包下载缓存。下次 flutter pub get 会重新下载。",
			SizePaths: []string{filepath.Join(home, ".pub-cache")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".pub-cache")}},
			Probe:     probePaths,
		},
		{
			ID: "huggingface", Group: "dev_caches", Title: "HuggingFace", Level: Costly,
			Hint:      "HuggingFace 模型与数据集缓存。下次加载会重新下载（GB 级）。",
			SizePaths: []string{filepath.Join(home, ".cache/huggingface")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".cache/huggingface")}},
			Probe:     probePaths,
		},
		{
			ID: "ollama", Group: "dev_caches", Title: "Ollama 模型", Level: Costly,
			Hint:      "Ollama 本地模型权重。删除后需 ollama pull 重新下载（每个模型 GB 级）。",
			SizePaths: []string{filepath.Join(home, ".ollama/models")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".ollama/models")}},
			Probe:     probePaths,
		},

		// ── ide ────────────────────────────────────────────────────────────
		{
			ID: "xcode_derived", Group: "ide", Title: "Xcode DerivedData", Level: Safe,
			Hint:      "Xcode 编译中间产物。下次构建会重新生成（首次较慢）。",
			SizePaths: []string{filepath.Join(home, "Library/Developer/Xcode/DerivedData")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Developer/Xcode/DerivedData")}},
			Probe:     probePaths,
		},
		{
			ID: "vscode_cache", Group: "ide", Title: "VSCode Cache", Level: Safe,
			Hint: "VSCode 缓存和日志。重启后自动重建。",
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
			Hint:      "JetBrains 系列 IDE（IntelliJ/PyCharm 等）的索引缓存。打开项目时会重建。",
			SizePaths: []string{filepath.Join(home, "Library/Caches/JetBrains")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/JetBrains")}},
			Probe:     probePaths,
		},
		{
			ID: "swift_pm", Group: "ide", Title: "Swift PM", Level: Safe,
			Hint:      "Swift Package Manager 下载缓存。下次构建会重新下载。",
			SizePaths: []string{filepath.Join(home, "Library/Caches/org.swift.swiftpm")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Caches/org.swift.swiftpm")}},
			Probe:     probePaths,
		},
		{
			ID: "xcode_archives", Group: "ide", Title: "Xcode Archives", Level: Destructive,
			Hint:      "Xcode 已归档的 .xcarchive。删除后无法重新符号化对应版本的崩溃日志。",
			SizePaths: []string{filepath.Join(home, "Library/Developer/Xcode/Archives")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Developer/Xcode/Archives")}},
			Probe:     probePaths,
		},

		// ── mobile ────────────────────────────────────────────────────────
		{
			ID: "ios_simulator_unavail", Group: "mobile", Title: "iOS Simulator 失效设备", Level: Safe,
			Hint:      "Xcode 不再支持的 iOS 模拟器版本。删除后无影响。",
			SizePaths: nil,
			Action:    Action{Kind: ActCmd, Program: "xcrun", Args: []string{"simctl", "delete", "unavailable"}},
			Probe:     probeCmd("xcrun"),
		},
		{
			ID: "ios_backup", Group: "mobile", Title: "iOS 设备备份", Level: Destructive,
			Hint:      "iTunes/Finder iOS 设备备份。删除后无法恢复对应快照。",
			SizePaths: []string{filepath.Join(home, "Library/Application Support/MobileSync/Backup")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Application Support/MobileSync/Backup")}},
			Probe:     probePaths,
		},
		{
			ID: "android_avd", Group: "mobile", Title: "Android AVD", Level: Costly,
			Hint:      "Android 模拟器镜像。删除后需重新创建/下载（GB 级）。",
			SizePaths: []string{filepath.Join(home, ".android/avd")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".android/avd")}},
			Probe:     probePaths,
		},

		// ── system ─────────────────────────────────────────────────────────
		{
			ID: "apfs_snapshots", Group: "system", Title: "APFS 本地快照", Level: Costly,
			Hint:      "Time Machine 在本地保留的临时快照。系统会自动重建。",
			SizePaths: nil,
			Action:    Action{Kind: ActCmd, Program: "tmutil", Args: []string{"thinlocalsnapshots", "/", "999999999", "1"}},
			Probe:     probeCmd("tmutil"),
		},
		{
			ID: "user_logs", Group: "system", Title: "用户日志", Level: Safe,
			Hint:      "~/Library/Logs 下的应用日志。系统和应用会按需重建。",
			SizePaths: []string{filepath.Join(home, "Library/Logs")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, "Library/Logs")}},
			Probe:     probePaths,
		},
		{
			ID: "system_logs_archived", Group: "system", Title: "已归档系统日志", Level: Safe,
			Hint:      "/private/var/log 下的 .gz/.bz2 归档日志。新日志写入不受影响。",
			SizePaths: []string{"/private/var/log"},
			Action: Action{
				Kind: ActRmGlobInDir,
				Dir:  "/private/var/log",
				Exts: []string{"gz", "bz2"},
			},
			Probe: probeWritableDir("/private/var/log"),
		},
		{
			ID: "quicklook_cache", Group: "system", Title: "QuickLook 缩略图", Level: Safe,
			Hint:      "QuickLook 预览缩略图缓存。系统会按需重建。",
			SizePaths: nil,
			Action:    Action{Kind: ActCmd, Program: "qlmanage", Args: []string{"-r", "cache"}},
			Probe:     probeCmd("qlmanage"),
		},
		{
			ID: "trash", Group: "system", Title: "回收站", Level: Destructive,
			Hint:      "~/.Trash 回收站。永久删除其中所有内容（无法恢复）。",
			SizePaths: []string{filepath.Join(home, ".Trash")},
			Action:    Action{Kind: ActRmDir, Paths: []string{filepath.Join(home, ".Trash")}},
			Probe:     probePaths,
		},
		{
			ID: "ds_store", Group: "system", Title: ".DS_Store 递归扫除", Level: Safe,
			Hint:      "递归删除 home 下所有 .DS_Store 文件。Finder 会按需重建。",
			SizePaths: []string{home},
			Action:    Action{Kind: ActDsStoreSweep, Paths: []string{home}},
			Probe:     probeAlways,
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

// MergeCustom appends custom items from config to the catalog.
func MergeCustom(items []Item, customs []config.CustomItem) []Item {
	existingIDs := make(map[string]bool)
	for _, item := range items {
		existingIDs[item.ID] = true
	}

	for _, c := range customs {
		id := c.ID
		if existingIDs[id] {
			id = "custom_" + id
		}

		level := parseConfigLevel(config.ParseLevel(c.Level))
		hint := c.Hint
		if hint == "" {
			hint = "自定义清理项：" + c.Path
		}

		expandedPath := sysutil.ExpandHome(c.Path)
		probePaths := func(item *Item) bool {
			for _, p := range item.SizePaths {
				if _, err := os.Stat(p); err == nil {
					return true
				}
			}
			return false
		}

		customItem := Item{
			ID:    id,
			Group: "custom",
			Title: c.Title,
			Hint:  hint,
			Level: level,
			SizePaths: []string{expandedPath},
			Action: Action{
				Kind:  ActRmDir,
				Paths: []string{expandedPath},
			},
			Probe: probePaths,
		}
		items = append(items, customItem)
		existingIDs[id] = true
	}

	return items
}

func parseConfigLevel(l config.Level) SafetyLevel {
	switch l {
	case config.LevelCostly:
		return Costly
	case config.LevelDestructive:
		return Destructive
	default:
		return Safe
	}
}
