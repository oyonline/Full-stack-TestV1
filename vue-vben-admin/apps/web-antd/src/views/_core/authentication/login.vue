<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';

import { computed, h, onMounted, ref } from 'vue';

import { AuthenticationLogin, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { message } from 'ant-design-vue';

import { useAuthStore } from '#/store';
import { getCaptchaApi } from '#/api';

defineOptions({ name: 'Login' });

const authStore = useAuthStore();
const loginFormRef = ref<any>();

// 验证码相关
const captchaBase64 = ref('');
const captchaId = ref('');

// 获取验证码
async function fetchCaptcha(clearCode = true) {
  try {
    const res = await getCaptchaApi();
    if (res.code === 200) {
      captchaBase64.value = res.data;
      captchaId.value = res.id;
      if (clearCode) {
        await loginFormRef.value?.getFormApi?.().setFieldValue('code', '');
      }
    }
  } catch (error) {
    console.error('获取验证码失败:', error);
  }
}

// 页面加载时获取验证码
onMounted(() => {
  fetchCaptcha(false);
});

// 验证码组件（修复：输入框+图片+文字 全部水平一行，宽度与上面一致）
const CaptchaInput = {
  props: ['modelValue'],
  emits: ['update:modelValue'],
  setup(
    props: { modelValue?: string },
    { emit }: { emit: (e: 'update:modelValue', value: string) => void },
  ) {
    return () =>
      h(
        'div',
        {
          class: 'flex items-center gap-2 w-full',
          style: 'height: 40px;',
        },
        [
          // 输入框：flex-1 自适应占满剩余空间
          h('input', {
            class:
              'h-full px-3 border border-gray-300 rounded focus:outline-none focus:border-blue-500 flex-1',
            placeholder: '请输入验证码',
            value: props.modelValue,
            onInput: (e: Event) => {
              emit('update:modelValue', (e.target as HTMLInputElement).value);
            },
          }),
          // 图片：固定宽度
          h('img', {
            src: captchaBase64.value,
            alt: '验证码',
            class:
              'h-full w-24 cursor-pointer border rounded bg-white object-contain flex-shrink-0',
            onClick: fetchCaptcha,
          }),
          // 刷新文字：压缩宽度，允许换行
          h(
            'span',
            {
              class:
                'text-xs text-gray-500 cursor-pointer hover:text-blue-500 text-center leading-tight flex-shrink-0',
              style: 'width: 40px;',
              onClick: fetchCaptcha,
            },
            ['看不清', h('br'), '刷新'],
          ),
        ],
      );
  },
};

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: $t('authentication.usernameTip'),
      },
      fieldName: 'username',
      label: $t('authentication.username'),
      rules: z.string().min(1, { message: $t('authentication.usernameTip') }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        placeholder: $t('authentication.passwordTip'),
      },
      fieldName: 'password',
      label: $t('authentication.password'),
      rules: z.string().min(1, { message: $t('authentication.passwordTip') }),
    },
    {
      component: CaptchaInput,
      fieldName: 'code',
      label: '验证码',
      rules: z.string().min(4, { message: '请输入4位验证码' }),
      defaultValue: '',
    },
  ];
});

// 自定义登录处理
async function handleLogin(values: any) {
  try {
    await authStore.authLogin({
      username: values.username,
      password: values.password,
      code: values.code,
      uuid: captchaId.value,
    });
  } catch (error: any) {
    message.error(error?.message || '登录失败，请重试');
    await fetchCaptcha();
  }
}
</script>

<template>
  <!-- 问题3、4、5：通过配置项隐藏其他登录方式 -->
  <AuthenticationLogin
    ref="loginFormRef"
    :form-schema="formSchema"
    :loading="authStore.loginLoading"
    :show-code-login="false"
    :show-qrcode-login="false"
    :show-register="false"
    :show-remember-me="true"
    :show-forget-password="false"
    :show-other-login="false"
    :show-third-party-login="false"
    @submit="handleLogin"
  />
</template>
