# Branding 系统说明

更新时间：2026-05-09

## 文档定位

记录当前已提交的「品牌 / Logo」全链路真相,覆盖三块:

1. 前端 logo fallback 链路(用户视角看到的破图兜底)。
2. 后端 `internal/branding` 包(渲染默认 logo PNG + 邮件模板嵌图)。
3. 动态 favicon 端点契约(浏览器 tab 图标如何回到后端兜底)。

来源:EPO-59 文档检查发现「branding 体系无独立文档」,EPO-65 收口。

---

## 1. 前端 fallback 链路

### 1.1 调用链总览

```
用户视角 (浏览器渲染 logo)
    │
    ▼
VbenLogo.vue                            ← 通用 logo 组件,所有出现 logo 的位置都用它
    │  props: src / srcDark / systemName / placeholderBgColor / fallbackOnError
    │  state: imgFailed (ref<boolean>)  ← <img @error> 触发后翻为 true
    │
    ├── (有 src 且 imgFailed=false) ─→ <img src=...> 正常渲染
    │
    └── (空 src 或 imgFailed=true)
            │
            ▼
        fallback 占位 <div>
            │
            ├── extractInitial(systemName || text)        ← 取首字
            └── pickForegroundColor(placeholderBgColor)    ← 按 WCAG 亮度选黑/白前景
```

`useSafeLogoUrl` 是 `@core/composables` 里另一个 URL 级 fallback 工具(下文 1.4 节),
**目前 VbenLogo 内部并未调用它**;两者是两套并存方案,不是同一条链。

### 1.2 关键文件 + 函数签名

| 角色 | 文件 | 关键导出 |
| --- | --- | --- |
| Logo 通用组件 | `vue-vben-admin/packages/@core/ui-kit/shadcn-ui/src/components/logo/logo.vue` | `defineOptions({ name: 'VbenLogo' })`,内部 `imgFailed = ref(false)`,`@error="imgFailed = true"` |
| Logo 组件 re-export | `vue-vben-admin/packages/@core/ui-kit/shadcn-ui/src/components/logo/index.ts` | `export { default as VbenLogo } from './logo.vue'` |
| 共享分享 utils(@core 层) | `vue-vben-admin/packages/@core/base/shared/src/utils/branding.ts` | `extractInitial(raw: string): string`<br>`pickForegroundColor(bgHex: string): '#1F2937' \| '#FFFFFF'` |
| App 层副本(antd 应用) | `vue-vben-admin/apps/web-antd/src/utils/branding.ts` | 同上两个函数(同实现,供 app 内 `*.test.ts` 直接 import) |
| URL fallback composable | `vue-vben-admin/packages/@core/composables/src/use-safe-logo-url.ts` | `useSafeLogoUrl(logoUrl, opts?)` 返回 `{ url, failed, onError }` |
| Composable 出口 | `vue-vben-admin/packages/@core/composables/src/index.ts` | `export * from './use-safe-logo-url'` |
| 单测(组件行为) | `vue-vben-admin/packages/@core/ui-kit/shadcn-ui/src/components/logo/__tests__/logo.test.ts` | `VbenLogo onerror behavior` |
| 单测(算法) | `vue-vben-admin/apps/web-antd/src/utils/branding.test.ts` | `extractInitial` / `pickForegroundColor` 单元用例 |
| 端到端 | `vue-vben-admin/apps/web-antd/e2e/logo-fallback.spec.ts` | 清空 logo 后破图首字色块 + favicon 回退断言 |

### 1.3 fallback 触发条件(`logo.vue:79-81` + 模板)

VbenLogo 进入「占位首字色块」分支当 **以下任一条件满足**:

| 条件 | 判定来源 | 触发原因 |
| --- | --- | --- |
| `src` 为空字符串(默认值)且 `srcDark` 也空 | `logoSrc` computed (`logo.vue:72-77`) | 后端 `sys_app_logo` 未配置 → `preferences.logo.source = ''` → 透传到 VbenLogo `:src=""` |
| `<img @error>` 触发 `imgFailed = true` 且 `fallbackOnError !== false` | `showFallback` computed (`logo.vue:79-81`) | 后端配置了 logo URL 但加载失败(404 / 跨域 / 网络) |
| 主题为 `dark` 且 `srcDark` 空,但 `src` 也空 | 同上 | 暗色主题没单独配 logo,且默认 logo 也缺失 |

`fallbackOnError` 默认 `true`(`logo.vue:59`),业务层无需显式开启。

#### 算法语义补充

- `extractInitial(raw)` 跳过空白与标点/符号(`\p{P}\p{S}`),遇到 CJK 字符直接保留,
  其余取首字母 `toUpperCase()`;全部不匹配时兜底返回 `'S'`(`branding.ts:10`)。
- `pickForegroundColor(bgHex)` 走 WCAG 相对亮度 (`0.2126R + 0.7152G + 0.0722B`),
  `L < 0.5` 返回 `#FFFFFF`,否则 `#1F2937`(`branding.ts:36-37`)。
  hex 输入支持 `#RGB` 与 `#RRGGBB`,非法输入降级返回 `#1F2937`。
- 注意:`@core/base/shared` 与 `apps/web-antd/src/utils/` 各有一份**几乎相同**的实现。
  TODO:这份重复实现长期看建议合并,见 §5。

### 1.4 `useSafeLogoUrl` 现状(`use-safe-logo-url.ts`)

```ts
export function useSafeLogoUrl(
  logoUrl: MaybeRef<string | null | undefined>,
  opts?: { fallbackUrl?: MaybeRef<string> },
): { url: ComputedRef<string>; failed: Ref<boolean>; onError: () => void };
```

- 行为:`url` 为空或 `failed=true` 时,降级到 `opts.fallbackUrl`。
- `logoUrl` 变化会自动 reset `failed`(`watch` 内部)。
- **当前没有任何组件 import 它**(grep 全仓只有自身导出),
  VbenLogo 走的是组件内 `imgFailed` 路线,并未通过 `useSafeLogoUrl` 拼 URL 兜底。
- TODO:若未来需要把 fallback 切换到「直接拿后端 default-logo URL 而非客户端绘色块」,
  这个 composable 是入口(把 `fallbackUrl` 接到 §3 的 `default-logo.png`)。当前未接,文档保留以备扩展。

### 1.5 调用方(谁在用 VbenLogo)

| 文件 | 用途 |
| --- | --- |
| `vue-vben-admin/packages/effects/layouts/src/basic/layout.vue:276 / :372` | 后台主框架 sidebar logo + 折叠态 logo |
| `vue-vben-admin/packages/effects/layouts/src/authentication/authentication.vue:96` | 登录/注册页大 logo |
| `vue-vben-admin/packages/effects/common-ui/src/components/index.ts:29` | 统一 re-export(全局可直接 `import { VbenLogo }`) |

入参链路:
- `preferences.logo.source` ← `sys_app_logo`(`apps/web-antd/src/utils/system-settings.ts:402`)
- `preferences.logo.placeholderBgColor` ← `sys_app_logo_placeholder_color`(`system-settings.ts:401`)
- `preferences.app.name` ← `sys_app_name`(`system-settings.ts:417`)→ 透传作为 `systemName` / `text`

---

## 2. 后端 `internal/branding` 包

包路径:`go-admin/internal/branding/`。

```
go-admin/internal/branding/
├── render.go        ← 默认 logo PNG 渲染 + ETag + LRU
├── render_test.go
├── mail.go          ← 邮件模板渲染(内嵌 base64 logo)
├── mail_test.go
└── fonts.go         ← 字体 embed 转发(实际 //go:embed 在 go-admin/assets/embed.go)
```

### 2.1 对外 API 列表

| 函数 | 文件 | 签名 | 用途 |
| --- | --- | --- | --- |
| `RenderDefaultLogoPNG` | `render.go:28` | `func RenderDefaultLogoPNG(text, bgHex string, size int) ([]byte, error)` | 渲染首字 fallback PNG。`text` 取首字符,`bgHex` 形如 `#RRGGBB`,`size` ∈ {16,32,64,96,128,256}。命中 LRU(256 entries)直接返回。 |
| `ETagFor` | `render.go:58` | `func ETagFor(text, bgHex string, size int, brandingSig string) string` | 由 `text|bgHex|size|brandingSig` 拼 sha1 取前 16 hex 作为强 ETag。`brandingSig` 由调用方塞入,通常映射为查询参数 `v`(版本信号,品牌切换后让客户端失效旧缓存)。 |
| `EmailLogoBase64` | `mail.go:13` | `func EmailLogoBase64(text, bgHex string) (string, error)` | 调 `RenderDefaultLogoPNG(text, bgHex, 96)`,封装为 `data:image/png;base64,...`,供 `<img src>` 内嵌。 |
| `RenderEmailTemplate` | `mail.go:41` | `func RenderEmailTemplate(appName, title, body, footer, bgHex string) (string, error)` | 渲染完整 HTML 邮件,首字 logo 由 `appName` 取首字 + `bgHex` 计算后 inline base64 嵌入。 |
| `EmailTemplateData` | `mail.go:30` | struct: `LogoDataURI template.URL; AppName, Title string; Body template.HTML; Footer string` | 邮件模板入参 schema(对外暴露,允许调用方手工拼模板)。 |
| `NotoSansSC` | `fonts.go:6` | `var NotoSansSC []byte` | 中文字体字节流(透传自 `assets.NotoSansSC`)。 |
| `InterSemiBold` | `fonts.go:9` | `var InterSemiBold []byte` | 英文字体字节流(透传自 `assets.InterSemiBold`)。 |

未导出但关键的内部函数:

| 内部函数 | 文件 | 作用 |
| --- | --- | --- |
| `parseTextChar` | `render.go:64` | 取 trim 后首个 rune;空串兜底 `'S'`。 |
| `parseHexColor` | `render.go:72` | 严格校验 `#RRGGBB` 七字符,失败返回 error。 |
| `renderPNG` | `render.go:84` | freetype 真渲染:CJK 用 NotoSansSC、其它用 InterSemiBold;字号 = `size * 0.55`,DPI=72,前景色由 `pickForeground` 决定。 |
| `pickForeground` | `render.go:134` | WCAG 亮度法,`L > 0.179` 选黑(`#000`),否则白(`#FFF`)。**与前端 `pickForegroundColor` 阈值不一致**(前端用 `L < 0.5`),见 §5 TODO。 |
| `lruGet` / `lruPut` | `render.go:165 / :178` | 256 entry sync.Mutex 保护的 slice LRU,key = `"text|bgHex|size"`。 |

### 2.2 路由注册

| 方法/路径 | 文件 | Handler |
| --- | --- | --- |
| `GET /api/v1/branding/default-logo.png` | `go-admin/app/admin/router/branding.go:17` | `apis.Branding.GetDefaultLogoPNG` |
| `GET /api/v1/branding/email-preview` | `go-admin/app/admin/router/branding.go:18` | `apis.Branding.GetEmailPreview` |

注册方式:`init()` 把 `registerBrandingRouter` 加入 `routerNoCheckRole`(`branding.go:9-11`),
即两个端点**不走 RBAC 校验**,任何登录态都可拉取(favicon / 邮件预览本就允许公开)。

Handler 入口:`go-admin/app/admin/apis/branding.go`:

| Handler | 关键参数 / 行为 |
| --- | --- |
| `GetDefaultLogoPNG` (`branding.go:52`) | query: `text`(默认 `S`)/ `bg`(默认 `#1d4ed8`)/ `size`(默认 64)/ `v`(版本)。<br>先算 `branding.ETagFor(...)` → 命中 `If-None-Match` 返回 304;<br>否则调 `RenderDefaultLogoPNG`,响应 `Cache-Control: public, max-age=3600` + `ETag`。 |
| `GetEmailPreview` (`branding.go:22`) | query: `appName` / `title` / `body` / `footer` / `bg`,调 `RenderEmailTemplate` 输出 `text/html`,响应 `Cache-Control: no-cache`。仅供 e2e / 调试预览。 |

### 2.3 字体 embed 机制

字体不是 `internal/branding` 自己 embed,而是统一收敛到 **`go-admin/assets`** 包再被 `branding/fonts.go` 转发引用。

| 文件 | 内容 |
| --- | --- |
| `go-admin/assets/embed.go` | `//go:embed fonts/NotoSansSC-Regular.ttf` → `var NotoSansSC []byte`<br>`//go:embed fonts/Inter-SemiBold.ttf` → `var InterSemiBold []byte` |
| `go-admin/assets/fonts/NotoSansSC-Regular.ttf` | 思源黑体简体常规字重,用于 CJK 字符渲染。 |
| `go-admin/assets/fonts/Inter-SemiBold.ttf` | Inter SemiBold,用于 ASCII / 拉丁字符。 |
| `go-admin/internal/branding/fonts.go` | 仅做 `var NotoSansSC = assets.NotoSansSC` / `var InterSemiBold = assets.InterSemiBold` 转发,目的是让 `internal/branding` 包的 import 树不跨出 `go-admin/`,且不直接依赖 `assets/` 路径。 |

字体如何被选择(`render.go:85-88`):

```go
fontData := InterSemiBold
if unicode.Is(unicode.Han, ch) {
    fontData = NotoSansSC
}
```

→ 单字符 ASCII / 拉丁走 Inter,中文(`unicode.Han`)走 Noto Sans SC。
韩文 / 假名 / 阿拉伯文等其它脚本目前都会落入 InterSemiBold,可能渲染豆腐块。
TODO:多脚本支持是后续工作,本期不做。

### 2.4 性能与内存

- **LRU 缓存**:同 `text|bgHex|size` 重复请求只渲染一次,256 entry 上限(`render.go:158`),
  容量按典型组合数(系统名 1 字 × 几种 bg × 6 种尺寸)远超实际需要。
- **字体常驻内存**:两份 TTF 字节流在 process 启动时即被 embed,GC 不回收(包级 var)。
- **freetype 解析**:每次未命中 LRU 都会 `truetype.Parse(fontData)` 重新构造 face;
  TODO:若未来证实热点,可以缓存 `*truetype.Font`,本期未做。

---

## 3. 动态 favicon 端点契约

### 3.1 端点

`GET /api/v1/branding/default-logo.png`(同 §2.2,favicon 与组件 fallback 共用同一端点)。

### 3.2 入参(query string)

| 参数 | 类型 | 必填 | 默认 | 说明 |
| --- | --- | --- | --- | --- |
| `text` | string | 否 | `S` | 单字符渲染文本。Handler 取 trim 后首个 rune;空 / 多字符多余部分截断。前端通常传 `encodeURIComponent(系统名首字大写)`。 |
| `bg` | string `#RRGGBB` | 否 | `#1d4ed8` | 背景色十六进制。非法输入返回 400。 |
| `size` | int | 否 | `64` | 仅允许 `{16, 32, 64, 96, 128, 256}`,其它返回 400。 |
| `v` | string | 否 | `""` | 版本/品牌签名,仅参与 ETag 计算,不影响图像内容。品牌切换或 logo 配置变更时调用方应同步换 `v` 让旧 ETag 失效。 |

### 3.3 返回

成功(200):

- `Content-Type: image/png`
- `Cache-Control: public, max-age=3600`(浏览器/CDN 缓存 1h)
- `ETag: "<sha1 前 16 hex>"` ← `branding.ETagFor(text, bg, size, v)`
- body:PNG 字节流

条件请求(304):
- 客户端带 `If-None-Match: <ETag>` 且与服务端当前 ETag 命中 → `204 No Content`(实为 `304 Not Modified`,见 `branding.go:65-68`)。

错误(400):
- size 不在白名单 / hex 不合法 / text 解析失败 → JSON `{"error": "..."}`。

### 3.4 客户端如何拼请求(`apps/web-antd/src/bootstrap.ts:24-47`)

```ts
function syncFavicon() {
  const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);
  const base = apiURL.replace(/\/$/, '');
  const logoSource = preferences.logo.source;
  const firstChar = encodeURIComponent(
    ([...(preferences.app.name || 'A')][0] ?? 'A').toUpperCase(),
  );
  const bg = encodeURIComponent(preferences.logo.placeholderBgColor ?? '#1d4ed8');

  const faviconURL = (size: number) =>
    logoSource
      ? logoSource
      : `${base}/v1/branding/default-logo.png?text=${firstChar}&bg=${bg}&size=${size}`;

  for (const [id, size] of [['favicon-16', 16], ['favicon-32', 32], ['favicon-180', 180]]) {
    const el = document.getElementById(id) as HTMLLinkElement | null;
    if (el) el.href = faviconURL(size);
  }
}
```

调用时机(`bootstrap.ts:101-102`):
1. `await syncAppConfigFromBackend()` ← 拉 `/api/v1/app-config`,把 `sys_app_logo` / `sys_app_logo_placeholder_color` / `sys_app_name` 写入 `preferences`。
2. `syncFavicon()` ← 立刻读 `preferences` 拼 favicon URL 写回 `<link>`。

逻辑总结:

- **后端配置了 `sys_app_logo`**(非空)→ favicon `href = preferences.logo.source`(直接用客户端 URL,不打 default-logo 端点)。
- **后端没配 `sys_app_logo`**(空)→ favicon `href = ${apiURL}/v1/branding/default-logo.png?text=<首字>&bg=<bg>&size=<size>`,由后端动态出图。

`index.html` 中三个 `<link>` 节点 id:`favicon-16` / `favicon-32` / `favicon-180`(`apps/web-antd/index.html:16-18`),
分别对应桌面 16px / 32px favicon 与 iOS apple-touch-icon 180px。

### 3.5 e2e 验证

`apps/web-antd/e2e/logo-fallback.spec.ts`:

- 用例「favicon href 含 `/api/v1/branding/default-logo.png`(无 logo 时回退 T5 接口)」(`spec.ts:121`):
  清空后端 `sys_app_logo` → 重启前端 → 断言三个 `<link>` 的 `href` 都包含 `default-logo.png` 路径。
- 用例「VbenLogo fallback div」(`spec.ts:82-90`):清 logo 后页面顶部 logo 位**不能**渲染 `<img>`,
  必须落到 fallback `<div>`(用样式类作为指纹)。

---

## 4. 端到端时序

```
浏览器开页面
   │
   ▼
index.html ─→ <link rel="icon" href="/favicon.ico">  ← 占位静态文件,首屏短暂出现
   │
   ▼
bootstrap.ts: syncAppConfigFromBackend()
   │   GET /api/v1/app-config (拉 sys_app_logo / sys_app_logo_placeholder_color / sys_app_name)
   │   → applySystemSettingsToRuntime → updatePreferences({ logo, app })
   │
   ▼
bootstrap.ts: syncFavicon()
   │   重写 <link id="favicon-{16,32,180}">.href
   │   ├─ logo 配置存在 → href = sys_app_logo
   │   └─ logo 配置空    → href = /api/v1/branding/default-logo.png?text=...&bg=...&size=...
   │                              │
   │                              ▼
   │                     后端 GetDefaultLogoPNG handler
   │                              │
   │                              ▼
   │                     branding.RenderDefaultLogoPNG(text, bg, size)
   │                              │
   │                              ├─ LRU hit  → 直接返回 cached PNG
   │                              └─ LRU miss → freetype 渲染 → 写 LRU → 返回
   │
   ▼
应用挂载 → 各处 VbenLogo 渲染
   │
   ├─ logoSrc 非空 + 加载成功 → <img>
   └─ logoSrc 空 / @error 触发 → 占位 <div>
                                  │
                                  ├─ extractInitial(systemName)
                                  └─ pickForegroundColor(placeholderBgColor)
```

---

## 5. 已知不一致 / TODO

| # | 现象 | 位置 | 影响 | 备注 |
| --- | --- | --- | --- | --- |
| 1 | `extractInitial` / `pickForegroundColor` 在 `@core/base/shared/src/utils/branding.ts` 与 `apps/web-antd/src/utils/branding.ts` 各有一份实现 | 见 §1.2 | 二份代码若一方修改不同步,会导致 fallback 视觉与单测预期偏差 | TODO:统一到 `@core/base/shared`,app 层只 re-export。 |
| 2 | 前端 `pickForegroundColor` 用阈值 `L < 0.5`,后端 `pickForeground` 用 `L > 0.179` | 前端 `branding.ts:36-37` / 后端 `render.go:144-148` | 同一品牌色在前端色块(JS)与后端 PNG(favicon / 邮件)上,可能选出**不同的前景色** | TODO:统一阈值与公式(后端 `pow 2.4` 写成 `* x` 简化近似亦不严谨),建议都按 WCAG 标准 `L > 0.179` 走。 |
| 3 | `useSafeLogoUrl` 已导出但无消费者 | `@core/composables/src/use-safe-logo-url.ts` | 当前没有"URL 级"fallback 通道,只有"组件内 imgFailed"通道 | TODO:若决定收敛为单通道,二选一删除或接入。 |
| 4 | 后端字体仅含 NotoSansSC + InterSemiBold | `assets/fonts/` | 韩 / 日 / 阿等首字会渲染豆腐 | 暂不支持,后续按需求扩 embed。 |
| 5 | 邮件模板 (`mail.go:70-108`) 与系统名 / 品牌色绑死 inline 模板 | 同上 | 主题切换或暗色邮件需求出现时需改模板 | 当前需求覆盖不到,保持简单。 |

---

## 6. 引用清单

前端:
- `vue-vben-admin/packages/@core/ui-kit/shadcn-ui/src/components/logo/logo.vue`
- `vue-vben-admin/packages/@core/base/shared/src/utils/branding.ts`
- `vue-vben-admin/apps/web-antd/src/utils/branding.ts`
- `vue-vben-admin/packages/@core/composables/src/use-safe-logo-url.ts`
- `vue-vben-admin/apps/web-antd/src/bootstrap.ts`
- `vue-vben-admin/apps/web-antd/index.html`
- `vue-vben-admin/apps/web-antd/src/utils/system-settings.ts`
- `vue-vben-admin/apps/web-antd/e2e/logo-fallback.spec.ts`

后端:
- `go-admin/internal/branding/render.go`
- `go-admin/internal/branding/mail.go`
- `go-admin/internal/branding/fonts.go`
- `go-admin/assets/embed.go`
- `go-admin/app/admin/apis/branding.go`
- `go-admin/app/admin/router/branding.go`
- `go-admin/app/admin/service/sys_config_keys.go`(`sys_app_logo_placeholder_color` 注册)
