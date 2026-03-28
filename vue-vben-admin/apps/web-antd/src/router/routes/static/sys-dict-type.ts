import type { RouteRecordRaw } from 'vue-router';

const BasicLayout = () => import('#/layouts/basic.vue');
const DictTypeDetail = () => import('#/views/admin/sys-dict-type/detail.vue');

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      hideInMenu: true,
      title: '字典类型详情',
    },
    path: '/admin/sys-dict-type/detail',
    children: [
      {
        name: 'SysDictTypeDetail',
        path: '',
        component: DictTypeDetail,
        meta: {
          hideInMenu: true,
          title: '字典类型详情',
        },
      },
    ],
  },
];

export default routes;
