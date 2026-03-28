import type { RouteRecordRaw } from 'vue-router';

const BasicLayout = () => import('#/layouts/basic.vue');
const RoleWorkspace = () => import('#/views/admin/sys-role/workspace.vue');

const routes: RouteRecordRaw[] = [
  {
    component: BasicLayout,
    meta: {
      hideInMenu: true,
      title: '新增角色',
    },
    path: '/admin/sys-role/create',
    children: [
      {
        name: 'SysRoleCreate',
        path: '',
        component: RoleWorkspace,
        meta: {
          hideInMenu: true,
          title: '新增角色',
        },
      },
    ],
  },
  {
    component: BasicLayout,
    meta: {
      hideInMenu: true,
      title: '编辑角色',
    },
    path: '/admin/sys-role/edit',
    children: [
      {
        name: 'SysRoleEdit',
        path: '',
        component: RoleWorkspace,
        meta: {
          hideInMenu: true,
          title: '编辑角色',
        },
      },
    ],
  },
];

export default routes;
