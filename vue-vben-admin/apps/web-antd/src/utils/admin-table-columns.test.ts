import { describe, expect, it } from 'vitest';

import {
  buildAdminTableColumnSettingsItems,
  buildAdminTableColumns,
  calculateAdminTableScrollX,
  createAdminTableColumnDefinitions,
  normalizeAdminTableColumnStates,
  reorderAdminTableColumnStates,
} from './admin-table-columns';

describe('admin-table-columns', () => {
  it('keeps system columns visible and fixed on the right', () => {
    const definitions = createAdminTableColumnDefinitions(
      [
        { key: 'name', title: '名称', width: 120 },
        { key: 'status', title: '状态', width: 100 },
        { fixed: 'right', key: 'action', title: '操作', width: 90 },
      ],
      {
        systemColumnKeys: ['action'],
      },
    );

    const states = normalizeAdminTableColumnStates(
      definitions,
      [
        {
          fixed: false,
          key: 'action',
          order: 0,
          visible: false,
          width: 900,
        },
      ],
      {
        systemColumnKeys: ['action'],
      },
    );

    expect(states.find((state) => state.key === 'action')).toMatchObject({
      fixed: 'right',
      visible: true,
      width: 600,
    });
  });

  it('merges stored state with new definitions and removes stale columns', () => {
    const definitions = createAdminTableColumnDefinitions(
      [
        { key: 'name', title: '名称', width: 120 },
        { key: 'status', title: '状态', width: 100 },
        { fixed: 'right', key: 'action', title: '操作', width: 90 },
      ],
      {
        systemColumnKeys: ['action'],
      },
    );

    const states = normalizeAdminTableColumnStates(
      definitions,
      [
        {
          fixed: false,
          key: 'legacy',
          order: 0,
          visible: true,
          width: 160,
        },
        {
          fixed: false,
          key: 'status',
          order: 1,
          visible: false,
          width: 140,
        },
        {
          fixed: false,
          key: 'name',
          order: 2,
          visible: true,
          width: 180,
        },
      ],
      {
        systemColumnKeys: ['action'],
      },
    );

    expect(states.map((state) => state.key)).toEqual(['status', 'name', 'action']);
    expect(states.find((state) => state.key === 'status')?.visible).toBe(false);
    expect(states.find((state) => state.key === 'name')?.width).toBe(180);
  });

  it('builds render columns and disables hiding the last business column', () => {
    const definitions = createAdminTableColumnDefinitions(
      [
        { key: 'name', title: '名称', width: 120 },
        { fixed: 'left', key: 'status', title: '状态', width: 100 },
        { fixed: 'right', key: 'action', title: '操作', width: 90 },
      ],
      {
        systemColumnKeys: ['action'],
      },
    );

    const states = normalizeAdminTableColumnStates(
      definitions,
      [
        {
          fixed: false,
          key: 'name',
          order: 0,
          visible: true,
          width: 140,
        },
        {
          fixed: 'left',
          key: 'status',
          order: 1,
          visible: false,
          width: 110,
        },
      ],
      {
        systemColumnKeys: ['action'],
      },
    );

    const columns = buildAdminTableColumns(definitions, states);
    const settingsItems = buildAdminTableColumnSettingsItems(definitions, states);

    expect(columns.map((column) => String(column.key))).toEqual(['name', 'action']);
    expect(calculateAdminTableScrollX(states)).toBe(254);
    expect(settingsItems.find((item) => item.key === 'name')?.disableVisibility).toBe(true);
  });

  it('reorders columns based on the current visible order', () => {
    const definitions = createAdminTableColumnDefinitions(
      [
        { fixed: 'left', key: 'status', title: '状态', width: 100 },
        { key: 'name', title: '名称', width: 120 },
        { key: 'dept', title: '部门', width: 120 },
        { fixed: 'right', key: 'action', title: '操作', width: 90 },
      ],
      {
        systemColumnKeys: ['action'],
      },
    );

    const states = normalizeAdminTableColumnStates(definitions, [], {
      systemColumnKeys: ['action'],
    });
    const nextStates = reorderAdminTableColumnStates(definitions, states, 2, 1);
    const columns = buildAdminTableColumns(definitions, nextStates);

    expect(columns.map((column) => String(column.key))).toEqual([
      'status',
      'dept',
      'name',
      'action',
    ]);
  });
});
