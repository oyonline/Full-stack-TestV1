<script lang="ts" setup>
/**
 * 系统管理 - 服务监控
 * 只读展示：GET /api/v1/server-monitor，Card + Descriptions
 */
import { onMounted, ref } from 'vue';

import { Button, Card, Descriptions } from 'ant-design-vue';

import { getServerMonitorApi } from '#/api/core';
import type { ServerMonitorInfo } from '#/api/core';

const loading = ref(false);
const errorMsg = ref('');
const data = ref<ServerMonitorInfo | null>(null);

async function fetchData() {
  loading.value = true;
  errorMsg.value = '';
  try {
    data.value = await getServerMonitorApi();
  } catch (e: unknown) {
    const err = e as { message?: string };
    errorMsg.value = err?.message ?? '加载失败';
    data.value = null;
  } finally {
    loading.value = false;
  }
}

function formatPercent(v: unknown): string {
  if (v != null && typeof v === 'number') return `${Number((v as number).toFixed(2))}%`;
  return '-';
}
function formatMb(v: unknown): string {
  if (v != null && typeof v === 'number') return `${v} MB`;
  return '-';
}
function formatGb(v: unknown): string {
  if (v != null && typeof v === 'number') return `${v} GB`;
  return '-';
}
function formatKbs(v: unknown): string {
  if (v != null && typeof v === 'number') return `${v} KB/s`;
  return '-';
}
function formatHours(v: unknown): string {
  if (v != null && typeof v === 'number') return `${v} 小时`;
  return '-';
}
function renderVal(v: unknown): string {
  if (v == null) return '-';
  if (typeof v === 'object') return JSON.stringify(v);
  return String(v);
}

onMounted(() => {
  fetchData();
});
</script>

<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-medium">服务监控</h2>
      <Button @click="fetchData">刷新</Button>
    </div>

    <div v-if="errorMsg" class="mb-4 text-red-600">
      {{ errorMsg }}
    </div>

    <div v-if="loading" class="py-8 text-center text-gray-400">
      加载中…
    </div>
    <template v-else-if="data">
      <Card class="mb-4" title="系统信息">
        <Descriptions :column="1" bordered size="small">
          <Descriptions.Item label="操作系统">{{ data.os?.goOs ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="架构">{{ data.os?.arch ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="编译器">{{ data.os?.compiler ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="Go 版本">{{ data.os?.version ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="协程数">{{ data.os?.numGoroutine ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="IP">{{ data.os?.ip ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="项目目录">{{ data.os?.projectDir ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="主机名">{{ data.os?.hostName ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="当前时间">{{ data.os?.time ?? '-' }}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card class="mb-4" title="内存">
        <Descriptions :column="1" bordered size="small">
          <Descriptions.Item label="已用">{{ formatMb(data.mem?.used) }}</Descriptions.Item>
          <Descriptions.Item label="总量">{{ formatMb(data.mem?.total) }}</Descriptions.Item>
          <Descriptions.Item label="使用率">{{ formatPercent(data.mem?.percent) }}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card class="mb-4" title="CPU">
        <Descriptions :column="1" bordered size="small">
          <Descriptions.Item label="核心数">{{ data.cpu?.cpuNum ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="使用率">{{ formatPercent(data.cpu?.percent) }}</Descriptions.Item>
          <Descriptions.Item label="详情">{{ renderVal(data.cpu?.cpuInfo) }}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card class="mb-4" title="磁盘">
        <Descriptions :column="1" bordered size="small">
          <Descriptions.Item label="总量">{{ formatGb(data.disk?.total) }}</Descriptions.Item>
          <Descriptions.Item label="已用">{{ formatGb(data.disk?.used) }}</Descriptions.Item>
          <Descriptions.Item label="使用率">{{ formatPercent(data.disk?.percent) }}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card class="mb-4" title="网络">
        <Descriptions :column="1" bordered size="small">
          <Descriptions.Item label="入站">{{ formatKbs(data.net?.in) }}</Descriptions.Item>
          <Descriptions.Item label="出站">{{ formatKbs(data.net?.out) }}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card class="mb-4" title="其他">
        <Descriptions :column="1" bordered size="small">
          <Descriptions.Item label="Swap 已用">{{ data.swap?.used ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="Swap 总量">{{ data.swap?.total ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="位置">{{ data.location ?? '-' }}</Descriptions.Item>
          <Descriptions.Item label="已运行">{{ formatHours(data.bootTime) }}</Descriptions.Item>
        </Descriptions>
      </Card>
    </template>
    <div v-else class="py-8 text-center text-gray-400">
      暂无数据
    </div>
  </div>
</template>
