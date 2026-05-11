<script lang="ts" setup>
import { computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Button, Empty } from 'ant-design-vue';

import AdminPageShell from '#/components/admin/page-shell.vue';

import SpuDetailContent from './components/SpuDetailContent.vue';

const route = useRoute();
const router = useRouter();

const spuId = computed(() => {
  const raw = route.params.id;
  const value = Array.isArray(raw) ? raw[0] : raw;
  const normalized = Number(value ?? 0);
  return Number.isFinite(normalized) && normalized > 0 ? normalized : 0;
});

function goBack() {
  void router.push('/admin/sys-spu');
}

watch(
  () => spuId.value,
  (id) => {
    if (!id) {
      // 无效 id 时由模板展示空态
    }
  },
  { immediate: true },
);
</script>

<template>
  <AdminPageShell header-mode="compact">
    <template #title>SPU 详情</template>
    <template #header-extra>
      <Button @click="goBack">返回列表</Button>
    </template>

    <div v-if="!spuId" class="app-radius-panel bg-white p-6 shadow-sm">
      <Empty description="缺少 SPU ID 参数">
        <Button type="primary" @click="goBack">返回列表</Button>
      </Empty>
    </div>

    <div v-else class="app-radius-panel bg-white p-4 shadow-sm md:p-5">
      <SpuDetailContent :spu-id="spuId" mode="page" readonly />
    </div>
  </AdminPageShell>
</template>
