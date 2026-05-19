# lt-clean

macOS 磁盘清理工具，专为开发者设计。TUI + CLI 双模式。

针对开发机上的缓存堆积：语言工具链（Homebrew、Go、npm、Cargo、pip…）、IDE 产物（Xcode DerivedData、JetBrains、VSCode）、移动端构建状态（iOS Simulator、Android SDK）、APFS 快照。

**不碰**浏览器缓存、照片库、邮件数据等用户应用状态。只清理开发者可以安全重建的内容。

单二进制，~5–10 MB，无运行时依赖。

## 安装

```bash
# go install
go install github.com/liaotuo/lt-clean@latest

# 或从 GitHub Releases 下载预编译版本
# 或 make build-small 从源码构建
```

## 用法

### TUI（默认）

```bash
lt-clean
```

交互式界面，`↑↓` 导航，`space` 选择，`c` 清理，`d` 干跑模式，`q` 退出。

### CLI

```bash
lt-clean scan                    # 扫描并打印表格
lt-clean scan --json             # JSON 输出

lt-clean clean --id brew,go_modcache    # 清理指定项
lt-clean clean --safe                   # 清理所有 Safe 级项目
lt-clean clean --safe --permanent       # 永久删除（不进回收站）
```

## 特性

- **默认进回收站** — 清理的文件移入 `~/.Trash`，可恢复
- **三级安全等级** — Safe（可安全清理）、Costly（重建较慢）、Destructive（用户数据，需确认）
- **可配置排除** — `~/.config/lt-clean/config.json` 排除特定项目

## 目录

36 项，分 4 组：

| 组          | 示例项目                                    |
|-------------|---------------------------------------------|
| dev_caches  | Homebrew, Go 模块, npm, Cargo, Docker       |
| ide         | Xcode DerivedData, VSCode, JetBrains        |
| mobile      | iOS Simulator, Android AVD                  |
| system      | APFS 快照, 用户日志, 回收站, .DS_Store      |

完整列表见 `lt-clean scan`。

## 平台

| 平台   | 状态 |
|--------|:----:|
| macOS  |  ✓   |
| Linux  |  ○   |
| Windows|  ○   |

## 开发

```bash
make test         # 单元测试
make build        # 调试构建
make build-small  # 发布构建（~2 MB）
```

## License

MIT
