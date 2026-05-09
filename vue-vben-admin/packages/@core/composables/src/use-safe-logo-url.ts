import type { MaybeRef } from 'vue';

import { computed, ref, unref, watch } from 'vue';

interface UseSafeLogoUrlOptions {
  fallbackUrl?: MaybeRef<string>;
}

export function useSafeLogoUrl(
  logoUrl: MaybeRef<string | null | undefined>,
  opts?: UseSafeLogoUrlOptions,
) {
  const failed = ref(false);

  watch(
    () => unref(logoUrl),
    () => {
      failed.value = false;
    },
  );

  const url = computed(() => {
    const src = unref(logoUrl);
    if (!src || failed.value) {
      return unref(opts?.fallbackUrl) ?? '';
    }
    return src;
  });

  const onError = () => {
    failed.value = true;
  };

  return { failed, onError, url };
}
