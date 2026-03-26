const BasicLayout = () => import('./basic.vue');
const AuthPageLayout = () => import('./auth.vue');
const RouteView = () => import('./route-view.vue');

const IFrameView = () => import('@vben/layouts').then((m) => m.IFrameView);

export { AuthPageLayout, BasicLayout, IFrameView, RouteView };
