import { reactive, ref } from 'vue';

import type { AdminPageResult, AdminTablePagination } from '#/utils/admin-crud';
import {
  createAdminTablePagination,
  resolveAdminErrorMessage,
} from '#/utils/admin-crud';

type QueryFactory<TQuery extends object> = () => TQuery;

interface UseAdminTableOptions<
  TItem,
  TQuery extends object,
  TParams extends Record<string, unknown>,
> {
  createParams: (query: TQuery) => TParams;
  createQuery: QueryFactory<TQuery>;
  fallbackErrorMessage: string;
  fetcher: (
    params: TParams & { pageIndex: number; pageSize: number },
  ) => Promise<AdminPageResult<TItem>>;
  pageSize?: number;
}

export function useAdminTable<
  TItem,
  TQuery extends object,
  TParams extends Record<string, unknown>,
>(options: UseAdminTableOptions<TItem, TQuery, TParams>) {
  const loading = ref(false);
  const tableData = ref<TItem[]>([]);
  const errorMsg = ref('');
  const query = reactive(options.createQuery()) as TQuery;
  const pagination = reactive<AdminTablePagination>(
    createAdminTablePagination(options.pageSize),
  );

  async function fetchList() {
    loading.value = true;
    errorMsg.value = '';

    try {
      const result = await options.fetcher({
        pageIndex: pagination.current,
        pageSize: pagination.pageSize,
        ...options.createParams(query),
      });

      tableData.value = result.list || [];
      pagination.total = result.count || 0;
    } catch (error) {
      errorMsg.value = resolveAdminErrorMessage(
        error,
        options.fallbackErrorMessage,
      );
      tableData.value = [];
      pagination.total = 0;
    } finally {
      loading.value = false;
    }
  }

  function onSearch() {
    pagination.current = 1;
    void fetchList();
  }

  function onReset() {
    Object.assign(query, options.createQuery());
    pagination.current = 1;
    void fetchList();
  }

  function onTableChange(pag: { current?: number; pageSize?: number }) {
    if (pag.current) {
      pagination.current = pag.current;
    }
    if (pag.pageSize) {
      pagination.pageSize = pag.pageSize;
    }
    void fetchList();
  }

  return {
    errorMsg,
    fetchList,
    loading,
    onReset,
    onSearch,
    onTableChange,
    pagination,
    query,
    tableData,
  };
}
