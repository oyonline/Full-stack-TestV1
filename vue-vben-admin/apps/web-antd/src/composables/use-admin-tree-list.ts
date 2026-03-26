import { reactive, ref } from 'vue';

import { resolveAdminErrorMessage } from '#/utils/admin-crud';

type QueryFactory<TQuery extends object> = () => TQuery;

interface UseAdminTreeListOptions<
  TItem,
  TQuery extends object,
  TParams extends Record<string, unknown>,
> {
  createParams: (query: TQuery) => TParams;
  createQuery: QueryFactory<TQuery>;
  fallbackErrorMessage: string;
  fetcher: (params: TParams) => Promise<TItem[]>;
}

export function useAdminTreeList<
  TItem,
  TQuery extends object,
  TParams extends Record<string, unknown>,
>(options: UseAdminTreeListOptions<TItem, TQuery, TParams>) {
  const loading = ref(false);
  const treeData = ref<TItem[]>([]);
  const errorMsg = ref('');
  const query = reactive(options.createQuery()) as TQuery;

  async function fetchList() {
    loading.value = true;
    errorMsg.value = '';

    try {
      const data = await options.fetcher(options.createParams(query));
      treeData.value = Array.isArray(data) ? data : [];
    } catch (error) {
      errorMsg.value = resolveAdminErrorMessage(
        error,
        options.fallbackErrorMessage,
      );
      treeData.value = [];
    } finally {
      loading.value = false;
    }
  }

  function onSearch() {
    void fetchList();
  }

  function onReset() {
    Object.assign(query, options.createQuery());
    void fetchList();
  }

  return {
    errorMsg,
    fetchList,
    loading,
    onReset,
    onSearch,
    query,
    treeData,
  };
}
