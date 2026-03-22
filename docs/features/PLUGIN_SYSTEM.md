# Chemistry UNO 插件开发指南

## 概述

Chemistry UNO 支持通过 `.cumod` 格式的插件文件为游戏添加自定义卡牌。`.cumod` 文件本质上是一个 **ZIP 压缩包**，包含插件的元数据和卡牌定义。

管理员可以在管理面板（`/admin/plugins`）上传和管理插件。

---

## 文件结构

```
my-plugin.cumod   (ZIP 压缩包)
├── manifest.json     ← 必须存在，插件元数据
└── cards.json        ← 可选，卡牌定义列表
└── client.js         ← 可选，客户端插件脚本（JS）
└── server.js         ← 可选，服务端插件脚本（JS）
```

### `manifest.json`（必须）

```json
{
  "name": "交换卡包",
  "version": "1.0.0",
  "author": "开发者名称",
  "description": "提供随机交换手牌的特殊卡牌",
  "game_version": "1.0",
  "scripts": {
    "client": "client.js",
    "server": "server.js"
  }
}
```

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| `name` | string | ✅ | 插件显示名称 |
| `version` | string | ✅ | 插件版本号（语义化版本，如 `1.0.0`） |
| `author` | string | ❌ | 作者名称 |
| `description` | string | ❌ | 插件功能描述 |
| `game_version` | string | ❌ | 兼容的游戏版本（仅供参考） |
| `script` | string | ❌ | 客户端脚本入口路径（兼容旧字段） |
| `scripts.client` | string | ❌ | 客户端脚本入口路径（如 `client.js`） |
| `scripts.server` | string | ❌ | 服务端脚本入口路径（如 `server.js`） |

> ⚠️ 若插件缺少 `manifest.json`，服务器将拒绝安装并返回错误。

---

### `cards.json`（可选）

定义插件引入的卡牌列表：

```json
[
  {
    "symbol": "SWAP3",
    "display_name": "三重交换",
    "effect_type": "swap",
    "effect_config": { "count": 3 },
    "default_count": 2,
    "color": "#06b6d4"
  },
  {
    "symbol": "FORCE2",
    "display_name": "强制二连",
    "effect_type": "force_play",
    "effect_config": { "count": 2 },
    "default_count": 2,
    "color": "#f97316"
  },
  {
    "symbol": "CVT24",
    "display_name": "二换四",
    "effect_type": "convert",
    "effect_config": { "source_count": 2, "target_count": 4 },
    "default_count": 1,
    "color": "#a855f7"
  }
]
```

---

## 卡牌字段说明

| 字段 | 类型 | 必须 | 说明 |
|------|------|------|------|
| `symbol` | string | ✅ | 卡牌唯一标识符（大写字母+数字，如 `SWAP3`），全局唯一 |
| `display_name` | string | ❌ | 游戏中显示的名称 |
| `effect_type` | string | ✅ | 效果类型，见下方 |
| `effect_config` | object | ✅ | 效果参数，JSON 对象 |
| `default_count` | int | ❌ | 默认牌组中的数量（默认 2） |
| `color` | string | ❌ | 前端显示颜色，HEX 格式（如 `#06b6d4`） |

---

## 效果类型（effect_type）

### `swap` — 随机交换手牌

将玩家手牌中 N 张随机牌放回摸牌堆，再摸 N 张新牌。

```json
{
  "effect_type": "swap",
  "effect_config": { "count": 3 }
}
```

| 参数 | 说明 |
|------|------|
| `count` | 交换的手牌数量（若手牌不足则交换全部） |

---

### `force_play` — 强制对手出牌

下一位玩家本轮必须额外打出 N 张牌，才能结束本回合。

```json
{
  "effect_type": "force_play",
  "effect_config": { "count": 2 }
}
```

| 参数 | 说明 |
|------|------|
| `count` | 强制额外出牌次数 |

> 💡 若对手无法出牌，可选择摸牌，此时强制出牌计数清零，回合正常推进。

---

### `convert` — 卡牌转换

消耗自身 N 张（包含已打出的这张共 N 张），从摸牌堆摸取 M 张新牌。

```json
{
  "effect_type": "convert",
  "effect_config": {
    "source_count": 2,
    "target_count": 4
  }
}
```

| 参数 | 说明 |
|------|------|
| `source_count` | 总共消耗的该卡数量（含已打出的 1 张） |
| `target_count` | 摸取的新牌数量 |

> ⚠️ 若手牌中该卡剩余数量不足（`source_count - 1` 张），则出牌失败。

---

## 制作步骤

1. **创建项目目录**：
   ```
   my-plugin/
   ├── manifest.json
   └── cards.json
   ```

2. **编写 `manifest.json`**（填写必要字段）。

3. **编写 `cards.json`**（可选，定义卡牌）。

4. **打包为 ZIP 并重命名为 `.cumod`**：

   **Windows（PowerShell）**：
   ```powershell
   Compress-Archive -Path manifest.json, cards.json, client.js, server.js -DestinationPath my-plugin.zip
   Rename-Item my-plugin.zip my-plugin.cumod
   ```

   **Linux / macOS**：
   ```bash
   cd my-plugin/
   zip -r ../my-plugin.cumod manifest.json cards.json client.js server.js
   ```

5. **上传安装**：在管理面板 `/admin/plugins` 点击「安装 .cumod」，上传文件。

---

## 注意事项

- `symbol` 字段在整个系统中必须唯一。若与已有卡牌冲突，安装会失败。
- 同一 `.cumod` 文件（以 SHA256 哈希判断）不能重复安装。
- 安装后需要点击「热重载」或重启服务器，让正在进行的游戏生效。
- 卡牌安装后可在 deck builder（自定义牌组）中选用。
- 删除插件将同时删除其所有卡牌，但不影响已开始的游戏。

---

## ????????

?????????? `client.js` ?????? `server.js`?

### ??????client.js?
????????????????????????? `(plugin, api, context)` ?????
- `plugin`??????
- `api`??????????????????`api.get / api.post / api.put / api.del / api.request`?
- `context`??? `console`?`storage`??? `onMessage`?????????????

```js
// client.js
api.get('/plugins').then(console.log)
context.onMessage((payload) => {
  console.log('server payload:', payload)
})
```

### ??????server.js?
???????????????????????????? `api.sendToAll / api.sendToRoom / api.sendToUser`
?????????`plugin_message`??

```js
// server.js
exports.onLoad = (ctx) => {
  ctx.api.sendToAll({ text: 'Hello from server plugin' })
}
```

---

## 调试建议

- 使用 `GET /api/plugins` 确认插件已正确安装并激活。
- 查看服务端日志（含 `[Plugin]` 标记）了解加载详情。
- 若效果未触发，检查 `effect_config` JSON 是否符合对应类型的格式。

---

## 示例插件

一个完整的示例插件文件可参考：[交换卡包示例](./examples/swap-pack/)

```json
// manifest.json
{
  "name": "交换卡包",
  "version": "1.0.0",
  "author": "Chemistry UNO Team",
  "description": "包含随机交换、强制出牌、二换四三种特殊效果卡牌"
}
```

```json
// cards.json
[
  {
    "symbol": "SWAP2",
    "display_name": "双重交换",
    "effect_type": "swap",
    "effect_config": { "count": 2 },
    "default_count": 2,
    "color": "#06b6d4"
  },
  {
    "symbol": "FORCE1",
    "display_name": "强制出牌",
    "effect_type": "force_play",
    "effect_config": { "count": 1 },
    "default_count": 2,
    "color": "#f97316"
  },
  {
    "symbol": "CVT13",
    "display_name": "一换三",
    "effect_type": "convert",
    "effect_config": { "source_count": 1, "target_count": 3 },
    "default_count": 2,
    "color": "#a855f7"
  }
]
```
