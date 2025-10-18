# specmuxer = 基于 tmux 的多 AI 编程工具编排器

最小目标是“可恢复、可观测、可扩展”
用途：一条命令在任意目录开工；重启后自动/手动恢复；status/top 一眼看清活跃度与树状关系；全部状态落在项目内。

1. run/resume/attach/status/top/stop/logs/gc/doctor 为基本命令；形如:
   - `specmuxer run codex -- <args>`
   - `resume <.|目录|--all>` (不传参时不会执行, 而是提供 help)
   - `attach <session>`
   - `status [--json|--wide]`
   - `top`
   - `stop <session|--all>`
   - `logs <session> [--follow]`
   - `gc [--dry-run]`
   - `doctor`
2. 保持后台运行与恢复：
   - 提供可 resume 恢复能力；以使得用户可以自行配置支持支持开机自动恢复（可选，通过系统服务）与手动恢复（二选一，默认手动）。
   - 重启后不但恢复 session 运行态, 也根据不同工具恢复其上下文, 比如找到这个目录对应的所有非用户杀掉的进程并调用 codex resume <id>
   - 恢复语义
     - user_killed=true 的会话不参与自动恢复
     - 异常退出/机器重启导致中断且 user_killed=false → resume 有资格尝试
     - 用户可以主动编辑 session 的 yml 以指定自定义恢复后是否触发指令, 比如 "请继续" "/goahead"
3. 项目内存储：所有配置与状态存放在 `.specmuxer/`，
   1. 包含 `conf.yml`, `stats.yml`
   2. `sessions/*.yml`
   3. logs/*.log；日志仅为会话日志，不做 WAL/事务, 默认单日志上限 200MB（可配）自动分片, 且自动按日滚动。
   4. `.specmuxer` 目录默认 0700、日志 0600；支持日志脱敏规则（regex）
4. session 命名与定位
   1. 以绝对路径生成稳定 project_id（路径哈希）
   2. tmux session 命名建议 `specmuxer:p-<hash>-<tool>-<name>-<create_time>`，避免冲突
5. 会话日志：每个 session 独立日志文件，逐行带时间戳；`logs <session> --follow` 等价 `tail -f`；支持容量/天数滚动策略（可选）。
6. 活跃度判断：
   - `status`/`top` 不仅看 tmux 会话，还判断工具是否“活着”（以“最近输出时间”为主；默认阈值 120s，可配 90/180 等）。
   - 树状呈现：以“目录（project）→ session（工具实例）→（pane 可选）”展示；支持 --json 输出供上层工具消费。
7. 原子封装: 兼容不同工具实, 要有适配器机制
   - 按工具定义 start/resume/health/stop/extract_state（声明式）；项目内可覆盖全局适配器；默认内置 codex/claude/...
   - 前面提到的 resume, 应该就包含了如何在 session 文件中保存可以唯一寻址的 id, 以及某一工具所支持的恢复功能
8. CLI 体验：
    - 支持 `-- <args>` 显式透传
    - 支持 `--env K:V` 的运行时环境指定, 该指定的环境变量会被透传到每一个 run 或 resume 的实际工具进程上
    - 由 specmuxer 代理的 cli 工具所有操作均可良好适配, 包括鼠标和触控板的滚动, 终端上的复制等
    - 由 specmuxer 代理的 cli 工具除了显示工具本体, 还显示 status bar 和 side panel
      - 当前 status bar 显示最近的活跃时间等，side panel 上显示一些快捷指令（bar 和 panel 支持 `session/*.yml` 中配置）
      - 这些附带的观测和干预功能当前只提供 MVP 体验, 是作为后续重要扩展项实现
    - 平台与依赖：支持 Linux/macOS（tmux ≥ 3.x）；Windows 暂不承诺（WSL 尽力）；不要求 root。
9. 工程质量
    - 额外 SLO（验收口径）：性能, 比如 status 在 ≤50 会话 P95 < 300ms；一次 resume 尝试 ≤20 会话 < 10s（受外部工具冷启动影响）；
    - GC 与接管：gc 可清理孤儿注册表/残留 tmux；发现“野生 tmux”可提示“接管（adopt）或忽略”。
    - 错误与自检：错误信息需指明项目/会话/适配器上下文并给出下一步建议；doctor 检查 tmux 版本、PATH、权限、配置语法等。
10. 开源质量：
    - README “5 分钟上手”；
    - /docs（concepts/adapters/recovery/faq）, 符合 godoc 规范, 配置 github.io 生成工作流；
    - SemVer 与变更日志；
    - MIT 许可；
    - Issue/PR 模板齐备;
11. 不做（本期）
    - 不做 WAL/事件回放
    - 不做远程多机调度
    - 不做云同步/遥测（可选项除外）
    - 不做前端界面