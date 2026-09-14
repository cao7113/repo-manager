# Repo Manager

管理本地 Git 仓库的命令行工具。当前 MVP 使用 YAML 保存追踪信息，使用本机 `git` 命令读取仓库状态。

## 安装运行

```bash
go install ./cmd/repo
repo --help
repo --version
```

开发版本显示为 `dev`，通过 GoReleaser 发布的二进制会显示对应的 Git tag 版本。

### 发布版本

版本号和 tag 使用本地 [Cocogitto](https://docs.cocogitto.io/) 管理，tag 格式为 `v1.2.3`。先执行检查，再选择版本增量：

```bash
task check
task release:patch  # 或 task release:minor / task release:major
task release:push
```

`cog bump` 会生成 changelog 提交并创建本地 `v*` tag；推送 tag 后，GitHub Actions 会自动调用 GoReleaser 创建 GitHub Release。也可以直接使用 `cog bump --version 1.2.3` 发布指定版本。

### 使用 mise 安装

发布版本可通过 mise 从 GitHub Releases 安装：

```bash
mise use -g 'github:cao7113/repo-manager'
repo --help
```

升级到最新版本：

```bash
mise upgrade repo-manager
```

也可以直接安装指定版本：

```bash
mise use -g 'github:cao7113/repo-manager@v0.1.0'
```

首次使用 mise 时，请先参考 [mise 安装文档](https://mise.jdx.dev/getting-started.html) 完成安装，并确保 mise 已加入 shell 环境。

默认配置文件是 `~/.config/repo-manager/repos.yaml`，也可以通过 `--config PATH` 或 `REPO_MANAGER_CONFIG` 指定。

## 命令

```bash
repo add /path/to/repo
repo clone https://github.com/example/project.git
repo clone https://github.com/example/project.git /path/to/project
repo clone all
repo ls [query]
repo l [query]
repo show repo-name-or-path
repo stat
```

`repo clone URL [PATH]` 直接遵循 `git clone URL [PATH]` 的行为：不提供目标路径时由 Git 按默认规则创建目录。clone 成功后会自动加入追踪。

`repo clone all` 只会为配置中路径不存在且有远端 URL 的条目执行 clone。`repo add` 不要求仓库配置 `origin`，没有远端时 `url` 留空。

配置示例：

```yaml
version: 1
repos:
  - name: project
    path: /Users/me/src/project
    url: https://github.com/example/project.git
    desc: Example project
    tags:
      - golang
```

## 参考

- https://github.com/alajmo/mani
- https://github.com/nosarthur/gita
- https://github.com/hakoerber/git-repo-manager
