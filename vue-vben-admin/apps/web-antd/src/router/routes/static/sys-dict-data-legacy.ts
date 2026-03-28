import type { RouteRecordRaw } from 'vue-router';

const BasicLayout = () => import('#/layouts/basic.vue');

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    path: '/admin/sys-dict-data',
    redirect: '/admin/sys-dict-type',
    meta: {
      hideInMenu: true,
      title: '字典数据旧入口',
    },
    children: [],
  },
];

export default routes;
