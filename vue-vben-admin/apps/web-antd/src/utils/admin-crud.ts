export interface AdminPageResult<T> {
  count: number;
  list: T[];
  pageIndex?: number;
  pageSize?: number;
}

export interface AdminTablePagination {
  current: number;
  pageSize: number;
  showSizeChanger: boolean;
  showTotal: (total: number) => string;
  total: number;
}

export function createAdminTablePagination(
  pageSize = 10,
): AdminTablePagination {
  return {
    current: 1,
    pageSize,
    total: 0,
    showSizeChanger: true,
    showTotal: (total: number) => `共 ${total} 条`,
  };
}

export function resolveAdminErrorMessage(
  error: unknown,
  fallback = '请求失败',
): string {
  const requestError = error as {
    message?: string;
    response?: { data?: { msg?: string } };
  };

  return requestError?.message || requestError?.response?.data?.msg || fallback;
}

export function renderAdminEmpty(
  value: null | number | string | undefined,
  fallback = '-',
): string {
  if (value === null || value === undefined || value === '') {
    return fallback;
  }
  return String(value);
}

export function formatAdminDateTime(
  value: null | string | undefined,
  fallback = '-',
): string {
  if (!value) {
    return fallback;
  }

  try {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }

    const pad = (num: number) => String(num).padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
  } catch {
    return value;
  }
}
