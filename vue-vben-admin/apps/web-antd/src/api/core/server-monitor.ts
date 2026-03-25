import { useAccessStore } from '@vben/stores';

import { getHttpRaw } from '#/api/request';

/** 系统信息（与 go-admin getOSInfo 对齐） */
export interface ServerMonitorOsInfo {
  goOs?: string;
  arch?: string;
  mem?: number;
  compiler?: string;
  version?: string;
  numGoroutine?: number;
  ip?: string;
  projectDir?: string;
  hostName?: string;
  time?: string;
}

/** 内存信息（与 go-admin getMemoryInfo 对齐，单位 MB） */
export interface ServerMonitorMemInfo {
  used?: number;
  total?: number;
  percent?: number;
}

/** CPU 信息（与 go-admin getCPUInfo 对齐） */
export interface ServerMonitorCpuInfo {
  cpuInfo?: unknown;
  percent?: number;
  cpuNum?: number;
}

/** 磁盘信息（与 go-admin getDiskInfo 对齐，total/used 单位 GB） */
export interface ServerMonitorDiskInfo {
  total?: number;
  used?: number;
  percent?: number;
}

/** 网络信息（与 go-admin getNetworkInfo 对齐，单位 KB/s） */
export interface ServerMonitorNetInfo {
  in?: number;
  out?: number;
}

/** Swap 信息（与 go-admin getSwapInfo 对齐） */
export interface ServerMonitorSwapInfo {
  used?: number;
  total?: number;
}

/** 服务监控接口返回（与 go-admin e.Custom(gin.H{...}) 对齐，整段 body） */
export interface ServerMonitorInfo {
  code?: number;
  os?: ServerMonitorOsInfo;
  mem?: ServerMonitorMemInfo;
  cpu?: ServerMonitorCpuInfo;
  disk?: ServerMonitorDiskInfo;
  net?: ServerMonitorNetInfo;
  swap?: ServerMonitorSwapInfo;
  location?: string;
  bootTime?: number;
}

/**
 * 获取服务监控信息
 * 对接 go-admin GET /api/v1/server-monitor（响应为顶层 code/os/mem/cpu/...，无 data 包装，故用 baseRequestClient 取整段 body 并手动带鉴权）
 */
export async function getServerMonitorApi(): Promise<ServerMonitorInfo> {
  const accessStore = useAccessStore();
  const token = accessStore.accessToken;
  const res = await getHttpRaw<ServerMonitorInfo>('/v1/server-monitor', {
    headers: {
      Authorization: token ? `Bearer ${token}` : '',
    },
  });
  return (res.data ?? {}) as ServerMonitorInfo;
}
