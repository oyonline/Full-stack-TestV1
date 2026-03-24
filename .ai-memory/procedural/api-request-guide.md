# 请求工具使用规范

## 概述

项目使用双客户端模式处理 API 请求：
1. `requestClient` - 自动提取 data（当前配置有问题，暂不使用）
2. `baseRequestClient` - 返回原始响应，手动处理

## 当前推荐做法（2026-03-23）

**统一使用 `baseRequestClient`，显式处理 `res.data`**

### 标准请求模式

```typescript
import { baseRequestClient } from '#/api/request';

// GET 请求
export async function getXxxApi(): Promise<DataType[]> {
  const res = await baseRequestClient.get('/v1/xxx');
  const responseData = res.data as any;
  
  if (responseData?.code === 200 && Array.isArray(responseData.data)) {
    return responseData.data;
  }
  
  console.warn('[Xxx API] Invalid response:', responseData);
  return [];
}

// POST 请求
export async function createXxxApi(data: CreateParams): Promise<boolean> {
  const res = await baseRequestClient.post('/v1/xxx', data);
  const responseData = res.data as any;
  
  return responseData?.code === 200;
}
```

## 响应结构处理

后端统一返回格式：
```json
{
  "code": 200,
  "data": [...],
  "requestId": "xxx",
  "msg": "success"
}
```

前端处理逻辑：
```typescript
const res = await baseRequestClient.get('/v1/xxx');
const responseData = res.data;  // 这是 {code, data, requestId, msg}

if (responseData?.code === 200) {
  return responseData.data;  // 返回业务数据
}
```

## 认证机制

`baseRequestClient` 已配置自动认证拦截器：
- 从 `useAccessStore` 动态获取 `accessToken`
- 自动添加 `Authorization: Bearer {token}` 请求头

无需手动处理认证。

## 环境配置

| 环境 | 配置文件 | API地址 |
|------|---------|---------|
| 开发 | `.env.development` | `http://172.16.97.127:10082/api` |
| 测试 | `.env.test`（新建） | `http://测试服务器IP:端口/api` |
| 生产 | `.env.production` | 生产环境地址 |

## 注意事项

1. **不要混用 `requestClient` 和 `baseRequestClient`**：当前 `requestClient` 配置有问题，统一用 `baseRequestClient`
2. **始终检查 `code === 200`**：后端业务错误也通过 code 返回
3. **类型断言后检查**：`as any` 后还要判断数据结构
4. **添加错误日志**：方便排查问题

## 相关文件

- 请求配置：`apps/web-antd/src/api/request.ts`
- API示例：`apps/web-antd/src/api/core/menu.ts`

---

*记录时间：2026-03-23*
