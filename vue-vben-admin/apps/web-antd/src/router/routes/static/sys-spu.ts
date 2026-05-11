import type { RouteRecordRaw } from 'vue-router';

const BasicLayout = () => import('#/layouts/basic.vue');
const SpuDetailPage = () => import('#/views/admin/sys-spu/detail-page.vue');

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      hideInBreadcrumb: false,
      hideInMenu: true,
      title: 'SPU 详情',
    },
    path: '/admin/sys-spu/detail/:id',
    children: [
      {
        name: 'SpuDetail',
        path: '',
        component: SpuDetailPage,
        meta: {
          hideInBreadcrumb: false,
          hideInMenu: true,
          title: 'SPU 详情',
        },
      },
    ],
  },
];

export default routes;
