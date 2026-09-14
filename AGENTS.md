# Prompt

本地有很多 git 仓库，需要统一管理，参考golang项目 https://github.com/alajmo/mani 编写自己的管理程序

repo命令支持：
repo add /path/to/a/repo    # 添加本地仓库追踪，校验是否是合法仓库和是否已经追踪
repo clone a-repo-url       # 委派给 git clone命令，并调用repo add加入追踪
repo clone all              # 检查仓库本地path是否存在，不存在就clone获取
repo ls                     # 列出追踪的仓库，支持按名字搜索
repo show repo-name-or-path # 展示 最总仓库 详情
repo stat                   # 整体追踪状态，是否有仓库处于未提交状态

RepoItem结构至少包括
- name
- path  # 本地仓库路径
- url   # 远端仓库地址
- desc
- tags

## 技术要点

- cli使用cobra