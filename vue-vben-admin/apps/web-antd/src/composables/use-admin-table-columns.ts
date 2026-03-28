import type { TableColumnType } from 'ant-design-vue';
import type { MaybeRefOrGetter } from 'vue';

import { computed, ref, toValue, watch } from 'vue';
import { useRoute } from 'vue-router';

import { useDebounceFn } from '@vueuse/core';
import { useUserStore } from '@vben/stores';
import { StorageManager } from '@vben/utils';

import {
  buildAdminTableColumns,
  buildAdminTableColumnSettingsItems,
  calculateAdminTableScrollX,
  createAdminTableColumnDefinitions,
  normalizeAdminTableColumnStates,
  reorderAdminTableColumnStates,
  resolveAdminTableColumnKey,
  resolveAdminTableColumnOptions,
} from '#/utils/admin-table-columns';
import type {
  AdminTableColumnFixed,
  AdminTableColumnOptions,
  AdminTableColumnState,
} from '#/utils/admin-table-columns';

interface AdminTableColumnStoragePayload {
  columns: AdminTableColumnState[];
  version: 1;
}

interface UseAdminTableColumnsOptions extends AdminTableColumnOptions {
  tableId: string;
}

const STORAGE = new StorageManager({
  prefix: 'admin-table-columns',
});

export function useAdminTableColumns<RecordType>(
  baseColumns: MaybeRefOrGetter<readonly TableColumnType<RecordType>[]>,
  options: UseAdminTableColumnsOptions,
) {
  const userStore = useUserStore();
  const route = useRoute();

  const mergedOptions = resolveAdminTableColumnOptions(options);
  const settingsOpen = ref(false);
  const states = ref<AdminTableColumnState[]>([]);
  const initialized = ref(false);

  const definitions = computed(() =>
    createAdminTableColumnDefinitions(toValue(baseColumns), mergedOptions),
  );
  const storageKey = computed(
    () =>
      `${userStore.userInfo?.userId ?? 'anonymous'}:${route.path}:${options.tableId}`,
  );

  function getStoredStates() {
    const payload = STORAGE.getItem<AdminTableColumnStoragePayload>(
      storageKey.value,
      null,
    );
    if (payload?.version !== 1 || !Array.isArray(payload.columns)) {
      return [];
    }
    return payload.columns;
  }

  function syncStates(nextStates?: AdminTableColumnState[]) {
    states.value = normalizeAdminTableColumnStates(
      definitions.value,
      nextStates ?? getStoredStates(),
      mergedOptions,
    );
  }

  const persistStates = useDebounceFn(() => {
    if (!initialized.value) {
      return;
    }
    STORAGE.setItem(storageKey.value, {
      columns: states.value,
      version: 1,
    } satisfies AdminTableColumnStoragePayload);
  }, 120);

  function updateStates(
    updater: (currentStates: AdminTableColumnState[]) => AdminTableColumnState[],
  ) {
    states.value = normalizeAdminTableColumnStates(
      definitions.value,
      updater(states.value),
      mergedOptions,
    );
  }

  function setColumnVisible(key: string, visible: boolean) {
    updateStates((currentStates) =>
      currentStates.map((state) =>
        state.key === key ? { ...state, visible } : state,
      ),
    );
  }

  function setColumnFixed(key: string, fixed: AdminTableColumnFixed) {
    updateStates((currentStates) =>
      currentStates.map((state) =>
        state.key === key ? { ...state, fixed } : state,
      ),
    );
  }

  function reorderColumns(oldIndex: number, newIndex: number) {
    updateStates((currentStates) =>
      reorderAdminTableColumnStates(
        definitions.value,
        currentStates,
        oldIndex,
        newIndex,
      ),
    );
  }

  function handleResizeColumn(width: number, column: TableColumnType<RecordType>) {
    const columnKey = resolveAdminTableColumnKey(column);
    if (!columnKey) {
      return;
    }
    updateStates((currentStates) =>
      currentStates.map((state) =>
        state.key === columnKey ? { ...state, width } : state,
      ),
    );
  }

  function restoreDefaultColumns() {
    STORAGE.removeItem(storageKey.value);
    syncStates([]);
  }

  watch(
    [definitions, storageKey],
    () => {
      initialized.value = false;
      syncStates();
      initialized.value = true;
    },
    {
      immediate: true,
    },
  );

  watch(
    states,
    () => {
      persistStates();
    },
    {
      deep: true,
    },
  );

  return {
    handleResizeColumn,
    reorderColumns,
    restoreDefaultColumns,
    scrollX: computed(() => calculateAdminTableScrollX(states.value, mergedOptions)),
    setColumnFixed,
    setColumnVisible,
    settingsColumns: computed(() =>
      buildAdminTableColumnSettingsItems(definitions.value, states.value),
    ),
    settingsOpen,
    tableColumns: computed(() =>
      buildAdminTableColumns(definitions.value, states.value),
    ),
  };
}
