import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const BasicLayout = () => import('#/layouts/basic.vue');

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      hideInMenu: true,
      title: $t('page.auth.profile'),
    },
    path: '/profile',
    children: [
      {
        name: 'Profile',
        path: '',
        component: () => import('#/views/_core/profile/index.vue'),
        meta: {
          hideInMenu: true,
          title: $t('page.auth.profile'),
        },
      },
    ],
  },
];

export default routes;
