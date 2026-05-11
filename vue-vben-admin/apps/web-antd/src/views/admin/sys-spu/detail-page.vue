<script lang="ts" setup>
import { useRoute, useRouter } from 'vue-router';

import { Button } from 'ant-design-vue';

import AdminPageShell from '#/components/admin/page-shell.vue';
import SpuDetailContent from './components/SpuDetailContent.vue';

const route = useRoute();
const router = useRouter();

const spuId = Number(route.params.id);

function goBack() {
  router.back();
}
</script>

<template>
  <AdminPageShell>
    <template #eyebrow>Product Center</template>
    <template #title>SPU 详情</template>
    <template #header-extra>
      <Button @click="goBack">返回</Button>
    </template>

    <SpuDetailContent
      v-if="!Number.isNaN(spuId) && spuId > 0"
      :spu-id="spuId"
      mode="page"
      :readonly="true"
    />
    <div v-else class="py-8 text-center text-sm text-slate-400">
      无效的 SPU ID
    </div>
  </AdminPageShell>
</template>
