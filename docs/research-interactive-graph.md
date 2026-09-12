# 交互式 OVSDB Schema 关系图方案调研

调研日期：2026-09-12。目标：替代当前 Graphviz 静态 PNG（边交叉多、无交互），实现"点击表节点 → 高亮所有直接和间接（传递闭包）相关表，弱化无关表"，并显著减少边交叉。

## 结论 / 建议

**推荐方案：Go 端输出 JSON 图数据 + 单个静态 HTML 页面，使用 Cytoscape.js + elkjs（通过 cytoscape-elk 插件做分层布局）。零构建步骤，两个 `<script>` 标签从 CDN 引入即可。**

理由：

1. **传递闭包高亮是 Cytoscape.js 的内置 API**：`node.successors()` / `node.predecessors()` 官方文档明确定义为"递归获取出边/入边可达的所有元素"（即传递闭包），点击高亮相关表只需一行代码，配合 `.addClass()` 弱化其余元素。
2. **ELK 的 `layered` 算法是目前 JS 生态中最强的 Sugiyama 风格分层布局**，专为"有方向的 node-link 图"设计，边交叉最小化策略比 Graphviz dot 更现代、可调；elkjs 提供 `elk.bundled.js` 单文件，可直接 `<script>` 引入，无需打包工具。
3. 两者均 MIT / EPL-2.0 许可，仓库维护活跃（cytoscape.js v3.34.3 发布于 2026-09-07）。
4. 备选：**AntV G6 v5**，内置 `hover-activate` 行为（`degree` + `inactiveState` 选项直接实现 N 跳邻居高亮 + 弱化），内置 dagre 布局，中文文档全，适合想少写代码的场景；但"传递闭包"语义不如 cytoscape 的 `successors()` 精确（degree 是跳数上限）。
5. 死路已排除：Mermaid 只有 click 回调、无高亮/过滤能力；d3-graphviz / @hpcc-js/wasm 只是"浏览器里的 Graphviz"，布局质量与现在一样差；SchemaSpy/DBML 类 ER 工具均不支持 OVSDB。

## 库对比

| 库 | 许可 | Stars / 维护状态（GitHub 实测） | 分层布局 | 邻居/传递高亮 | 备注 |
|---|---|---|---|---|---|
| **Cytoscape.js** | MIT | 11.2k；最近 push 2026-09-10，v3.34.3 (2026-09-07) | 经插件：cytoscape-elk (ELK layered)、cytoscape-dagre、fCoSE、cola | **内置**：`successors()`/`predecessors()` 递归（传递闭包）；`neighborhood()` 仅直连；`bfs()`/`dfs()` | 图论 API 最全，官方定位即"visualisation and analysis" |
| **AntV G6 v5** | MIT | 12.3k；5.1.1 (2026-04-17)，push 2026-07-15 | 内置 DagreLayout / AntVDagreLayout | 内置行为 `hover-activate`（`degree` 跳数 + `direction` + `inactiveState` 弱化），`click-select` | 开箱交互最多，React/Vue 无关 |
| **vis-network** | Apache-2.0 | 3.6k；v10.1.2 (2026-08-19) | 内置 hierarchical（基于自己的算法） | 无内置，需手动遍历 edges | 简单场景可用，高亮要手写 |
| **Sigma.js** (+graphology) | MIT | 12.2k；push 2026-08-20 | 无分层布局（force-atlas2 等力导向） | graphology 有 BFS，但需自己接渲染 | WebGL 面向万级节点，本场景过度 |
| **React Flow (xyflow)** | MIT | 38.3k；维护极活跃 | **官方明确无内置布局**，文档推荐 dagre / elkjs（有官方 elkjs 示例） | 需自己算 + setNodes | 需要 React 技术栈，重；官方文档原话 "We have not implemented our own layouting solution yet" |
| **Mermaid** | MIT | 90k；极活跃 | flowchart (dagre) / elk 可选 | **无任何内置高亮/过滤**；click 回调需 `securityLevel='loose'` | 是"绘图语言"不是图分析库，交互靠回调自己写 DOM 操作，不适合 |
| **d3-graphviz** (@hpcc-js/wasm) | BSD-3 / Apache-2.0 | 1.8k / 389；wasm 版 2026-09 仍更新 | Graphviz dot 本身（交叉问题原样保留） | 无 | 只解决"浏览器渲染 DOT"，不解决交叉，高亮需手动操作 SVG |

布局/插件附属仓库状态：elkjs（kieler/elkjs，2.7k stars，push 2026-09-09，EPL-2.0）、cytoscape-elk（57 stars，v2.3.0，MIT）、cytoscape.js-fcose（183 stars，push 2026-04-17，MIT）、cytoscape.js-dagre（287 stars，push 2026-08-28）、dagre（5.8k stars，2025-11 复活发布 v2.0.0）。

## 布局算法结论（边交叉最小化）

- **ELK `layered`**：Sugiyama 框架的现代实现，官方描述"particularly suited for node-link diagrams with an inherent direction and ports"，基于 Sugiyama et al. 的方法。支持 port 概念（边可锚定到字段级），分层 + 交叉最小化 + 边路由全套，是 ER/schema 图的最优选择。elkjs 默认包含 `'layered', 'stress', 'mrtree', 'radial', 'force', 'disco'`。
- **dagre**：同为 Sugiyama 风格，轻量、"largely a drop-in solution"（React Flow 官方语），质量略逊于 ELK 但够用；2025-11 发布 v2.0.0 恢复维护。
- **fCoSE / cola（cytoscape 插件）**：力导向带约束，适合无向图，对 ER 这类有向层次图交叉控制不如 layered。
- **Graphviz via WASM**：布局算法就是现在的 dot，换渲染端不解决任何问题，排除。
- 实务建议：ER 图 50–200 节点，ELK layered 与 dagre 都能秒出；ELK 可调项多（`elk.direction`、`elk.layered.crossingMinimization.*`），优先 ELK。

## 传递闭包高亮（官方 API 引证）

Cytoscape.js 官方文档原文：

- `nodes.successors()` — "Recursively get edges (and their targets) coming out of the nodes in the collection (i.e. the outgoers, the outgoers' outgoers, …)."
- `nodes.predecessors()` — "Recursively get edges (and their sources) coming into the nodes in the collection (i.e. the incomers, the incomers' incomers, …)."
- `eles.neighborhood()` — 仅"directly connected elements"。

因此"点击表 → 高亮上下游传递闭包 + 弱化其余"：

```js
cy.on('tap', 'node', (evt) => {
  const n = evt.target;
  const related = n.successors().union(n.predecessors()).union(n);
  cy.elements().addClass('dimmed');
  related.removeClass('dimmed').addClass('highlight');
});
cy.on('tap', (e) => { if (e.target === cy) cy.elements().removeClass('dimmed highlight'); });
```

G6 v5 等价物：`behaviors: [{ type: 'hover-activate', degree: 1, direction: 'both', inactiveState: 'inactive' }]`（官方配置表确认 `degree` 控制激活跳数、`inactiveState` 赋给未激活元素；degree 是跳数上限而非完整传递闭包语义，但 `degree` 传足够大即可覆盖）。

## 架构选型

| 选项 | 结论 |
|---|---|
| a. Go 输出 JSON + 静态 HTML + JS 库 | **推荐**。改动最小：复用现有 RFC 7047 schema 解析（`ovsdb/` 包），新增一个 JSON 序列化输出；前端一个 `index.html`，CDN 引入 cytoscape/elk/cytoscape-elk，无 npm 无构建 |
| b. 转给现成 ER 工具（SchemaSpy/DBML/dbeaver） | 排除。SchemaSpy 走 JDBC、DBML 需手写中间格式，均不支持 OVSDB JSON schema，等于再写一遍转换器还得不到传递高亮 |
| c. 浏览器内 Graphviz（d3-graphviz / @hpcc-js/wasm） | 排除。布局即 dot，交叉问题原样保留 |

### 最小落地计划

1. **Go 端**：`main.go` 增加 `-format json` 输出（或新子命令），把现有解析结果（表名、列、uuid 引用 → 边）序列化为：
   ```json
   {"nodes":[{"id":"Logical_Switch","data":{"columns":[...]}}],
    "edges":[{"id":"e1","source":"Logical_Switch","target":"Port","label":"ports"}]}
   ```
   现有 graphviz 输出保留不动，纯增量。
2. **前端单文件 `web/index.html`**（也可 `go:embed` 进二进制单文件分发）：
   ```html
   <script src="https://unpkg.com/cytoscape@3/dist/cytoscape.min.js"></script>
   <script src="https://unpkg.com/elkjs@0.12.0/lib/elk.bundled.js"></script>
   <script src="https://unpkg.com/cytoscape-elk@2.3.0/dist/cytoscape-elk.js"></script>
   ```
   布局配置：`layout: { name: 'elk', elk: { 'elk.algorithm': 'layered', 'elk.direction': 'DOWN' } }`。
3. **交互**：上面的 tap 处理器 + 两条 CSS class（`.dimmed { opacity: 0.15 }`、`.highlight` 描边变色）。
4. **取舍说明**：Graphviz 版节点是 HTML-like record（列出所有列名），Cytoscape 原生节点只支持文本 label。最小方案节点只显示表名，列信息放点击侧边栏或 tooltip（数据已在 JSON 里）；若坚持字段级 record 外观，可加 cytoscape-node-html-label 插件，非必需。

工作量估计：Go 侧 ~100 行，HTML 侧 ~150 行。

## Sources

- Cytoscape.js 仓库状态：https://github.com/cytoscape/cytoscape.js （GitHub API 实测 stars/pushed/release）
- Cytoscape.js 遍历 API 文档（successors/predecessors/neighborhood 原文）：https://js.cytoscape.org/#eles.successors
- elkjs 仓库、算法列表、elk.bundled.js 用法、EPL-2.0：https://github.com/kieler/elkjs
- ELK layered 算法（Sugiyama）说明：https://www.eclipse.org/elk/reference/algorithms/org-eclipse-elk-layered.html
- cytoscape-elk 插件：https://github.com/cytoscape/cytoscape.js-elk 、https://www.npmjs.com/package/cytoscape-elk
- cytoscape fCoSE / dagre / cola 插件：https://github.com/iVis-at-Bilkent/cytoscape.js-fcose 、https://github.com/cytoscape/cytoscape.js-dagre 、https://github.com/cytoscape/cytoscape.js-cola
- dagre v2 复活：https://github.com/dagrejs/dagre
- AntV G6 仓库：https://github.com/antvis/G6 ；行为总览：https://g6.antv.antgroup.com/en/manual/behavior/overview ；hover-activate 配置（degree/direction/inactiveState）：https://g6.antv.antgroup.com/en/manual/behavior/hover-activate ；布局总览（DagreLayout/AntVDagreLayout）：https://g6.antv.antgroup.com/en/manual/layout/overview
- React Flow 官方 layouting 文档（"We have not implemented our own layouting solution yet"，推荐 dagre/elkjs）：https://reactflow.dev/learn/layouting/layouting
- Mermaid 交互限制（click 回调、securityLevel='loose'、无高亮/过滤）：https://mermaid.js.org/syntax/flowchart.html#interaction
- vis-network：https://github.com/visjs/vis-network ；Sigma.js：https://github.com/jacomyal/sigma.js ；graphology：https://github.com/graphology/graphology
- d3-graphviz：https://github.com/magjac/d3-graphviz ；@hpcc-js/wasm：https://github.com/hpcc-systems/hpcc-js-wasm
- SchemaSpy 文档（JDBC-only）：https://schemaspy.readthedocs.io/ ；DBML：https://dbml.dbdiagram.io/home/ （均无 OVSDB 支持）
- RFC 7047（OVSDB 协议/schema 格式依据）：https://www.rfc-editor.org/rfc/rfc7047

---

# 追加调研（2026-09-12 第二轮）：ER 工具路线 vs dot 就地改进

## 角度 A：OVSDB schema → 标准数据库 schema → 复用 ER 图工具

### A.1 类型系统转换的坑（RFC 7047 原文核对）

RFC 7047 §3.2 定义的类型系统与 SQL 差异很大：

- **uuid + refTable → FK**：自然映射。但 `refType` 分 `strong`/`weak`（§3.2，"refType": "strong" or "weak"，仅随 refTable 出现）；weak ref 允许悬空（"If refType is weak, then any UUIDs are allowed"，且弱引用为空时被自动清理，§5.x）。SQL FK 只有强引用语义，**weak/strong 区别在 ER 图里无法表达，除非人工标注**。
- **set/map 无一等表达**：OVSDB 的列值可以是 set（`<base-type>` 数组 + min/max 基数约束）或 map（key/value 两个 base-type，如 `external_ids: map<string,string>`）。SQL 没有多值列，转 DDL 只能把 set<uuid> 拆成关联表（人为制造大量"中间实体"节点，图反而更乱）或降级为文本列（**丢掉引用边，图直接出错**）。
- **ephemeral 列**（§3.2 "ephemeral": 值不持久化）：SQL 无对应概念。
- **isRoot**：GC 根标记（§3.2 根表不被垃圾回收），ER 图无对应概念，只能视觉标注。

结论：转换器要正确处理"set<uuid> refTable=X" 这种最常见的引用形式，本质还是要自己解析 RFC 7047 schema 并提取"表→表"引用关系——**这正是现有 Go 代码已经在做的事**，转 SQL/DBML 是绕路且丢信息。

### A.2 离线 ER 渲染工具逐个核实

| 工具 | 输入 | 离线渲染 | 布局引擎（实测） | 交互能力（官方文档核实） |
|---|---|---|---|---|
| **dbdiagram.io / @dbml/core + dbml-renderer** | DBML 文本 | dbml-renderer CLI 本地出 SVG | **viz.js（即 Graphviz dot 的 JS 移植）**，见 package.json 依赖 `@aduh95/viz.js` | 输出**静态 SVG**，README 无任何交互功能；dbdiagram.io 是 Web 服务，官方 DBML 文档无 hover 高亮关联的说明 |
| **Mermaid erDiagram** | 文本 | 可（mermaid CLI/JS 本地跑） | **v12.0.0 起默认改用 ELK**（官方文档："are laid out by ELK rather than Dagre"，ELK "suits larger and more-complex diagrams"），可切回 dagre | erDiagram 文档**没有 interaction 章节**（flowchart 才有 click 回调）；无实体点击高亮、无关联过滤。想交互只能自己 post-process 输出的 SVG |
| **PlantUML IE/ER 图** | 文本 | 可（本地 jar） | **依赖 Graphviz dot**（官方：class 系图需要 Graphviz；Smetana 只是 dot 的纯 Java 移植，算法相同） | SVG 支持链接/提示，无邻居高亮类交互 |
| **eralchemy** | SQLAlchemy 模型或**活数据库连接** | 可，但输出静态 PNG/SVG/PDF | graphviz/pygraphviz | 无 |
| **SchemaCrawler** | **JDBC 活连接**（离线 snapshot 需先从活库序列化） | — | graphviz | 无交互式高亮 |
| **DBeaver** | **数据库连接**（从 Connections 里双击表/schema 打开 Diagram） | — | 内置 | 有 ER 浏览但需先建连接；无"隐藏无关表"功能文档 |
| **drawio SQL 插件** | DDL 文本导入 | 可 | 手工/简单自动布局 | 手动拖拽，无传递关联高亮 |

OVSDB 不是 SQL 数据库，起不了实例，需要活连接的工具（SchemaCrawler/DBeaver 常规路径）直接出局；吃文本的工具（DBML/Mermaid/PlantUML）布局又回到 Graphviz 或只能静态出图。**唯一亮点是 Mermaid v12 的 erDiagram 默认 ELK 布局**，但交互为零。

### A.3 路线 A 总成本评估

写一个 OVSDB→DBML 转换器 ≈ 150-250 行（解析已有，加映射 + 绕开 set/map/weak ref 的表达难题），换来的最好结果是 Mermaid ELK 静态图（布局改善、交互为零）或 dbml-renderer 的 dot 静态图（什么都没改善）。信息损失（weak ref、多值列）直接影响图的正确性。**路线 A 否决。**

## 角度 B：dot/Graphviz 就地改进

### B.1 布局调优手段（Graphviz 官方文档核对）

- **`tred`**：官方定义为"Transitive reduction filter for directed graphs"——去掉对可达性冗余的边。对引用密集图能减边，但**有语义代价**：若 A→B、B→C、A→C 都是真实引用，tred 会删掉 A→C 这条真实存在的 FK 边。可用于"骨架视图"，不能作为默认。
- **`unflatten`**：官方定位"Adjust directed graphs to improve layout aspect ratio"（通过插入不可见节点/边把过宽的图拉窄），改善的是宽高比和可读性，间接减少拥挤。选项细节见其 man page。
- **dot 的交叉最小化强度参数**：`remincross`（默认 true，有多 cluster 时二次跑交叉最小化，仅 dot 有效）、`mclimit`/`nslimit`（迭代次数上限，调大可多花 CPU 换更少交叉）、`searchsize`。
- **其他常用手段**：`rankdir=LR`（宽图改横向分层）、`splines=ortho`（正交折线，视觉上交叉更规整）/ `splines=curved`、`concentrate=true`（合并平行边）、`newrank`（改进 rank 算法）、`ordering`/`invis` 不可见边引导排序（值得试但属于手工微调）。
- **换引擎**（官方定位）：dot = "hierarchical or layered drawings of directed graphs"，本来就是这类图的正确选择；neato/fdp/sfdp 是 spring/force-directed，ER 层次图只会更乱；twopi radial、circo circular、osage/patchwork 面向 cluster，均不适合。**引擎不用换，问题在 dot 的 Sugiyama 实现本身不如 ELK。**

**B.1 结论**：上述手段组合（`rankdir=LR` + `splines=ortho` + `unflatten -l` + 调大 `mclimit`/`searchsize` + 可选 `concentrate`）能明显改善现状，估计减少两三成视觉交叉，但上限是 dot 算法本身，达不到 ELK layered 的水平。

### B.2 dot 输出的交互化

| 方案 | 核实结果 |
|---|---|
| **xdot**（jrfonseca/xdot.py） | Python3 + GTK3 + Cairo，939 stars，push 2026-03-19，LGPL-3.0，维护中。功能：任意缩放、"Highlights node/edge **under mouse**"（只高亮鼠标下的单个元素，**不是邻居**）、节点 URL 事件、点击边聚焦源/目标节点。**做不到传递闭包高亮**。 |
| **d3-graphviz**（+ @hpcc-js/wasm） | 浏览器内 wasm 版 Graphviz 布局出 SVG。官方 README 确认：生成的 SVG 结构与普通 SVG 一致、可用 d3 选择操作（`attributer` 钩子逐元素生效、`keyMode('id')` 用 DOT 节点 id、`<title>` 含 DOT 名称）；内置 pan/zoom（`graphviz.zoom()`）与 transition；自定义点击处理用标准 d3（`end` 事件后 `selectAll('g.node').on('click', ...)`，官方有 Delete Nodes 交互 demo）。**可以完全不动 Go 端 dot 输出，另附一份 edges JSON，在 JS 里算传递闭包做高亮/隐藏。** |
| 其他离线交互 dot viewer | 基本凋零（ZGRViewer 等多年不维护），Gephi 可导入 DOT 但交互靠手动操作，不适合"点一下出闭包"的诉求。 |

### B.2 结论

dot 路线用尽手段后：**布局最多"明显改善"而非"解决交叉"；交互唯一可行解是 d3-graphviz + 自写 JS 闭包高亮**（技术上完全可行，约 100 行 JS）。

## 三方终评与最终建议

| | A. ER 工具 | B1. dot 调优+静态 | B2. d3-graphviz | C1. Cytoscape+ELK | **C2. elkjs + 原生 SVG** |
|---|---|---|---|---|---|
| 交叉少 | ✗（dot/静态） | ~ | ✗（dot） | ✓ | ✓ |
| 点击高亮传递闭包+隐藏无关 | ✗ | ✗ | ✓ | ✓ | ✓ |
| Go 端改动 | 大（转换器） | 零 | 零（加 edges JSON） | 小（JSON 输出） | 小（JSON 输出） |
| 前端复杂度 | — | — | 1 个 HTML + ~100 行 JS | cytoscape API | 1 个 HTML + ~200 行 vanilla JS |
| 依赖 | 第三方服务/工具链 | graphviz | 2 个 CDN 脚本 | 3 个 CDN 脚本 | **1 个 CDN 脚本（elk.bundled.js）** |

**最终建议：C2 —— elkjs 单库 + 手写原生 SVG 渲染。**

理由：用户嫌 Cytoscape 复杂，而本需求其实不需要图编辑/样式系统/插件生态，只需要"布局 + 画矩形和折线 + 点击事件"。elkjs 的 `layout()` 输出就是节点 x/y/宽高 + 边的 bend points 数组，渲染成 SVG 是纯粹的机械代码（节点画 `<rect>`+`<text>`，边画 `<polyline>`/`marker` 箭头），加一个 `click` 处理器用邻接表 BFS 出传递闭包、给无关元素加 `opacity:0.15` 即可。全部 vanilla JS 无框架、单 CDN 依赖、布局质量是 ELK layered（比 dot 交叉更少）。"隐藏无关"用 CSS class 切换，比任何库都快。

- 若要**绝对最小改动**（Go 端一行不动）：选 B2（d3-graphviz），接受 dot 级布局，配 B1 的调优参数。作为快速验证交互原型也合适。
- Cytoscape 方案保留为"以后想要更多图分析交互（搜索、路径、过滤面板）"时的升级路径。
- A 路线（ER 工具）正式否决：转换丢语义、最好结果也只是静态图。

### 本轮 Sources

- RFC 7047 §3.2（refTable/refType/ephemeral/isRoot）、§5.1（set/map 编码）：https://www.rfc-editor.org/rfc/rfc7047
- Mermaid erDiagram 文档（v12 起默认 ELK 布局、无 interaction 章节）：https://mermaid.js.org/syntax/entityRelationshipDiagram.html
- dbml-renderer 仓库（viz.js 依赖、静态 SVG 输出）：https://github.com/softwaretechnik-berlin/dbml-renderer 、其 package.json 中 `@aduh95/viz.js`
- DBML 官方文档（无高亮交互说明）：https://dbml.dbdiagram.io/docs/
- PlantUML 与 Graphviz 依赖关系（class 系图需 Graphviz、Smetana 为 Java 移植）：https://plantuml.com/graphviz-dot
- eralchemy（"from databases or from SQLAlchemy models"、依赖 graphviz/pygraphviz）：https://github.com/eralchemy/eralchemy
- SchemaCrawler（JDBC 活连接模型）：https://www.schemacrawler.com/
- DBeaver ER 图（从 Connections 打开，面向已配置连接）：https://dbeaver.com/docs/dbeaver/ER-Diagrams/
- Graphviz 官方文档：tred https://graphviz.org/docs/cli/tred/ 、unflatten https://graphviz.org/docs/cli/unflatten/ 、remincross https://graphviz.org/docs/attrs/remincross/ 、布局引擎定位 https://graphviz.org/docs/layouts/
- xdot.py 功能与维护状态：https://github.com/jrfonseca/xdot.py
- d3-graphviz README（SVG 可控性、zoom/transition、自定义 click）：https://github.com/magjac/d3-graphviz
