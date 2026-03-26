import { computed } from 'vue';

import { useAccess } from '@vben/access';
import { useAccessStore } from '@vben/stores';

interface UseAdminPermissionOptions {
  codes?: string | string[];
  roles?: string | string[];
}

function normalizeValues(value?: string | string[]) {
  if (!value) {
    return [];
  }

  return (Array.isArray(value) ? value : [value]).filter(Boolean);
}

function matchesPermission(granted: string, required: string) {
  if (!granted || !required) {
    return false;
  }
  if (granted === '*' || granted === '*:*:*') {
    return true;
  }
  if (granted === required) {
    return true;
  }

  const grantedSegments = granted.split(':');
  const requiredSegments = required.split(':');
  const maxLength = Math.max(grantedSegments.length, requiredSegments.length);

  for (let index = 0; index < maxLength; index += 1) {
    const grantedSegment = grantedSegments[index];
    const requiredSegment = requiredSegments[index];

    if (grantedSegment === '*') {
      return true;
    }
    if (grantedSegment == null || requiredSegment == null) {
      return false;
    }
    if (grantedSegment !== requiredSegment) {
      return false;
    }
  }

  return true;
}

export function useAdminPermission(options: UseAdminPermissionOptions = {}) {
  const { hasAccessByCodes, hasAccessByRoles } = useAccess();
  const accessStore = useAccessStore();

  const codeList = computed(() => normalizeValues(options.codes));
  const roleList = computed(() => normalizeValues(options.roles));
  const grantedCodes = computed(() => normalizeValues(accessStore.accessCodes));

  const hasPermission = computed(() => {
    const codePass =
      codeList.value.length === 0 ||
      hasAccessByCodes(codeList.value) ||
      codeList.value.some((requiredCode) =>
        grantedCodes.value.some((grantedCode) =>
          matchesPermission(grantedCode, requiredCode),
        ),
      );
    const rolePass =
      roleList.value.length === 0 || hasAccessByRoles(roleList.value);
    return codePass && rolePass;
  });

  return {
    hasPermission,
  };
}
