# Repo Manager

管理本地 Git 仓库的命令行工具。当前 MVP 使用 YAML 保存追踪信息，使用本机 `git` 命令读取仓库状态。

## 安装运行

```bash
go install ./cmd/repo
repo --help
```

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
