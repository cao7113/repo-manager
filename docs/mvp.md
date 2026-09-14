# Repo Manager MVP 技术方案

## 1. 背景与目标

本项目用于统一管理本地多个 Git 仓库，参考 `mani`、`gita` 等工具，提供统一的仓库登记、克隆、查询和状态检查能力。

MVP 的目标是建立一个可运行的命令行工具，解决以下问题：

- 记录本地仓库的位置和基本信息
- 校验并追踪已有的本地 Git 仓库
- 使用标准 `git clone` 行为获取新仓库
- 快速列出和查看已追踪仓库
- 汇总所有仓库当前是否存在未提交变更

MVP 不负责执行提交、推送、拉取或批量修改仓库内容。

## 2. 设计决策

### 2.1 CLI 框架

使用 [Cobra](https://github.com/spf13/cobra) 构建命令树：

- 根命令：`repo`
- 全局参数：`--config PATH`
- 子命令：`add`、`clone`、`ls`、`show`、`stat`

Cobra 负责参数校验、帮助信息和子命令分发，业务逻辑保留在独立函数和内部包中。

### 2.2 Git 调用方式

不引入 Git 库，统一调用本机 `git` 可执行文件：

- 仓库校验：`git rev-parse --show-toplevel`
- 远端查询：`git remote get-url origin`
- 克隆：`git clone ...`
- 状态查询：`git status --porcelain=v2 --branch`

这样可以保持与用户本机 Git 配置和行为一致，并避免维护 Git 协议实现。

### 2.3 clone 行为

```bash
repo clone URL
repo clone URL PATH
```

以上命令直接对应：

```bash
git clone URL
git clone URL PATH
```

不提供目标路径时，由 Git 按自身规则决定默认目录名。clone 成功后，程序自动调用 add 逻辑将仓库加入追踪。

批量克隆使用：

```bash
repo clone all
```

该命令只处理配置中本地路径不存在、且存在远端 URL 的仓库。已经存在的路径会跳过，没有 URL 的仓库会提示并跳过。

### 2.4 无 origin 仓库

`repo add` 只要求路径是合法 Git 仓库，不要求存在 `origin` 远端。

没有 `origin` 时：

- 仓库仍然可以被添加
- `url` 字段为空
- `repo clone all` 会跳过该仓库，因为无法自动克隆

## 3. 命令规格

### 3.1 `repo add PATH`

添加一个已有的本地 Git 仓库。

处理流程：

1. 展开 `~` 并转换为绝对路径。
2. 使用 `git rev-parse --show-toplevel` 定位仓库根目录。
3. 将仓库根目录规范化后作为 `path` 保存。
4. 使用仓库根目录名作为默认 `name`。
5. 尝试读取 `origin` URL；读取失败时保存空 URL。
6. 校验路径和名称是否重复。
7. 原子写入配置文件。

重复路径或重复名称会返回错误，不会覆盖原有记录。

### 3.2 `repo clone URL [PATH]`

调用 `git clone`，成功后将结果加入配置。

如果 clone 成功但自动 add 失败，命令会返回错误；已创建的本地 Git 仓库不会被自动删除，便于用户检查和恢复。

### 3.3 `repo clone all`

遍历所有配置项：

- `path` 已存在：跳过
- `url` 为空：跳过并提示
- `path` 不存在且有 URL：创建父目录并执行 clone

MVP 采用串行处理，不自动执行 fetch，也不提供并发参数。

### 3.4 `repo ls [QUERY]`

列出已追踪仓库。没有查询条件时列出全部仓库；有查询条件时，对仓库名称和本地路径进行不区分大小写的包含匹配。

当前输出为面向终端阅读的两列文本：

```text
NAME                 PATH
```

### 3.5 `repo show NAME_OR_PATH`

按名称或路径查找单个仓库，并展示：

- name
- path
- url
- desc
- tags
- 当前分支
- 当前状态

状态信息在执行命令时实时读取，不完全依赖配置文件。

### 3.6 `repo stat`

遍历所有追踪仓库并输出状态摘要。

状态包括：

- `clean`：工作区干净
- `modified`：存在已跟踪文件修改
- `modified, untracked`：存在修改和未追踪文件
- `conflicted`：存在冲突
- `ahead`：本地领先上游
- `behind`：本地落后上游
- `diverged`：本地和上游均有独立提交
- `missing/invalid`：路径不存在或不是可用 Git 仓库

`stat` 不执行网络操作，因此不会隐式执行 `git fetch`。

## 4. 数据模型

配置文件使用 YAML，默认位置为：

```text
~/.config/repo-manager/repos.yaml
```

也可以通过以下方式覆盖：

```bash
repo --config /path/to/repos.yaml ls
REPO_MANAGER_CONFIG=/path/to/repos.yaml repo ls
```

数据结构：

```yaml
version: 1
repos:
  - name: repo-manager
    path: /Users/me/src/repo-manager
    url: https://github.com/example/repo-manager.git
    desc: Local repository manager
    tags:
      - golang
      - cli
```

对应 Go 结构：

```go
type RepoItem struct {
    Name string   `yaml:"name"`
    Path string   `yaml:"path"`
    URL  string   `yaml:"url,omitempty"`
    Desc string   `yaml:"desc,omitempty"`
    Tags []string `yaml:"tags,omitempty"`
}
```

`version` 用于未来升级配置格式。路径统一保存为绝对路径，名称默认取 Git 仓库根目录名。

## 5. 模块划分

```text
cmd/repo/
    main.go                 # Cobra 根命令、子命令和流程编排

internal/config/
    config.go               # 配置模型、路径解析、YAML 读写

internal/git/
    git.go                  # git 命令调用和 porcelain v2 状态解析

internal/repository/
    repository.go            # 仓库添加、查找、搜索和重复校验
```

职责边界：

- `cmd/repo` 处理 CLI 参数、输出和命令流程。
- `internal/config` 只负责配置文件和 `RepoItem` 数据模型。
- `internal/git` 封装外部 Git 命令，不让命令层直接拼接 Git 调用。
- `internal/repository` 处理与 Git 无关的仓库集合操作。

## 6. 配置写入策略

保存配置时使用临时文件再重命名覆盖目标文件：

1. 创建配置目录。
2. 序列化 YAML。
3. 在同一目录创建临时文件。
4. 写入临时文件并关闭。
5. 使用 `rename` 替换正式配置文件。

这样可以避免进程中断时直接留下半截配置文件。当前 MVP 尚未实现跨进程文件锁；后续若支持并发运行，应增加锁或重试机制。

## 7. 测试策略

### 单元测试

重点覆盖不依赖真实 Git 的逻辑：

- 重复路径校验
- 重复名称校验
- 名称搜索
- 路径搜索
- 配置读写和默认版本
- porcelain v2 状态解析

### 集成测试

后续可在临时目录中创建真实 Git 仓库，覆盖：

- 添加无 `origin` 仓库
- 添加有 `origin` 仓库
- clone 后自动追踪
- `clone all` 补齐缺失仓库
- 工作区修改、未追踪文件和冲突状态

## 8. 当前限制

MVP 暂不包含：

- 删除追踪项
- 编辑 desc 和 tags 的专用命令
- JSON 或其他机器可读输出
- 自动 fetch
- 并行 clone 或 stat
- 配置文件锁
- 仓库 URL 规范化
- 多仓库批量 pull、push、checkout

这些能力可以在不改变核心 `RepoItem` 模型和 Git 封装边界的前提下逐步增加。

## 9. 验证方式

在项目根目录执行：

```bash
go test ./...
go vet ./...
go run ./cmd/repo --help
```

安装后可以使用：

```bash
go install ./cmd/repo
repo --help
```
