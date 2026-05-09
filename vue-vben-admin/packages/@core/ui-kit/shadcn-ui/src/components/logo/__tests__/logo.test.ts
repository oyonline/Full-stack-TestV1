import { mount } from '@vue/test-utils';

import { describe, expect, it } from 'vitest';

import VbenLogo from '../logo.vue';

describe('VbenLogo onerror behavior', () => {
  it('renders img when valid logoSrc is provided', () => {
    const wrapper = mount(VbenLogo, {
      props: {
        src: 'https://example.com/logo.png',
        text: 'Vben',
      },
    });

    expect(wrapper.find('img').exists()).toBe(true);
  });

  it('shows fallback div after img error event', async () => {
    const wrapper = mount(VbenLogo, {
      props: {
        src: 'https://example.com/logo.png',
        text: 'Vben',
        fallbackOnError: true,
      },
    });

    expect(wrapper.find('img').exists()).toBe(true);

    await wrapper.find('img').trigger('error');

    expect(wrapper.find('img').exists()).toBe(false);
    expect(wrapper.find('div.flex.items-center.justify-center').exists()).toBe(true);
  });

  it('keeps img visible after error when fallbackOnError is false', async () => {
    const wrapper = mount(VbenLogo, {
      props: {
        src: 'https://example.com/logo.png',
        text: 'Vben',
        fallbackOnError: false,
      },
    });

    expect(wrapper.find('img').exists()).toBe(true);

    await wrapper.find('img').trigger('error');

    expect(wrapper.find('img').exists()).toBe(true);
  });
});
