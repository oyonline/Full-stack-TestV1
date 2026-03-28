import type { TableColumnType } from 'ant-design-vue';

export type AdminTableColumnFixed = false | 'left' | 'right';

export interface AdminTableColumnState {
  fixed: AdminTableColumnFixed;
  key: string;
  order: number;
  visible: boolean;
  width: number;
}

export interface AdminTableColumnDefinition<RecordType = any> {
  column: TableColumnType<RecordType>;
  defaultFixed: AdminTableColumnFixed;
  defaultOrder: number;
  defaultWidth: number;
  key: string;
  system: boolean;
  title: string;
}

export interface AdminTableColumnSettingsItem extends AdminTableColumnState {
  disableFixed: boolean;
  disableVisibility: boolean;
  draggable: boolean;
  system: boolean;
  title: string;
}

export interface AdminTableColumnOptions {
  defaultWidth?: number;
  maxWidth?: number;
  minWidth?: number;
  scrollBuffer?: number;
  systemColumnKeys?: string[];
}

const DEFAULT_WIDTH = 160;
const MAX_WIDTH = 600;
const MIN_WIDTH = 80;
const SCROLL_BUFFER = 24;

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

function coerceColumnWidth(value: number | string | undefined, fallback: number) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === 'string') {
    const parsed = Number.parseFloat(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return fallback;
}

function getColumnOrderRank(
  state: AdminTableColumnState,
  definition: AdminTableColumnDefinition | undefined,
) {
  if (definition?.system) {
    return 3;
  }
  if (state.fixed === 'left') {
    return 0;
  }
  if (state.fixed === 'right') {
    return 2;
  }
  return 1;
}

function toColumnKey<RecordType>(
  column: TableColumnType<RecordType>,
  fallbackIndex: number,
) {
  if (column.key != null) {
    return String(column.key);
  }
  if (typeof column.dataIndex === 'number' || typeof column.dataIndex === 'string') {
    return String(column.dataIndex);
  }
  return `__column_${fallbackIndex}`;
}

function toColumnTitle<RecordType>(
  column: TableColumnType<RecordType>,
  fallbackKey: string,
) {
  if (typeof column.title === 'number' || typeof column.title === 'string') {
    return String(column.title);
  }
  return fallbackKey;
}

function normalizeFixedValue(value: AdminTableColumnFixed | boolean | undefined) {
  if (value === true || value === 'left') {
    return 'left' as const;
  }
  if (value === 'right') {
    return 'right' as const;
  }
  return false;
}

function toDefinitionsMap<RecordType>(
  definitions: AdminTableColumnDefinition<RecordType>[],
) {
  return new Map(definitions.map((definition) => [definition.key, definition]));
}

function sortStatesForRender<RecordType>(
  definitions: AdminTableColumnDefinition<RecordType>[],
  states: AdminTableColumnState[],
) {
  const definitionsMap = toDefinitionsMap(definitions);
  return [...states].sort((left, right) => {
    const leftDefinition = definitionsMap.get(left.key);
    const rightDefinition = definitionsMap.get(right.key);
    const leftRank = getColumnOrderRank(left, leftDefinition);
    const rightRank = getColumnOrderRank(right, rightDefinition);
    if (leftRank !== rightRank) {
      return leftRank - rightRank;
    }
    if (left.order !== right.order) {
      return left.order - right.order;
    }
    return (
      (leftDefinition?.defaultOrder ?? Number.MAX_SAFE_INTEGER) -
      (rightDefinition?.defaultOrder ?? Number.MAX_SAFE_INTEGER)
    );
  });
}

function normalizeSingleState(
  definition: AdminTableColumnDefinition,
  state: AdminTableColumnState | undefined,
  options: Required<AdminTableColumnOptions>,
): AdminTableColumnState {
  const nextState: AdminTableColumnState = {
    fixed: normalizeFixedValue(state?.fixed ?? definition.defaultFixed),
    key: definition.key,
    order: typeof state?.order === 'number' ? state.order : definition.defaultOrder,
    visible: state?.visible ?? true,
    width: clamp(
      coerceColumnWidth(state?.width, definition.defaultWidth),
      options.minWidth,
      options.maxWidth,
    ),
  };

  if (definition.system) {
    nextState.fixed = 'right';
    nextState.visible = true;
  }

  return nextState;
}

export function createAdminTableColumnDefinitions<RecordType>(
  baseColumns: readonly TableColumnType<RecordType>[],
  options: AdminTableColumnOptions = {},
): AdminTableColumnDefinition<RecordType>[] {
  const mergedOptions = resolveAdminTableColumnOptions(options);
  const systemColumnKeySet = new Set(mergedOptions.systemColumnKeys);

  return baseColumns.map((column, index) => {
    const key = toColumnKey(column, index);
    return {
      column,
      defaultFixed: systemColumnKeySet.has(key)
        ? 'right'
        : normalizeFixedValue(column.fixed as AdminTableColumnFixed | boolean),
      defaultOrder: index,
      defaultWidth: clamp(
        coerceColumnWidth(column.width as number | string | undefined, mergedOptions.defaultWidth),
        mergedOptions.minWidth,
        mergedOptions.maxWidth,
      ),
      key,
      system: systemColumnKeySet.has(key),
      title: toColumnTitle(column, key),
    };
  });
}

export function normalizeAdminTableColumnStates<RecordType>(
  definitions: AdminTableColumnDefinition<RecordType>[],
  states: AdminTableColumnState[] = [],
  options: AdminTableColumnOptions = {},
) {
  const mergedOptions = resolveAdminTableColumnOptions(options);
  const stateMap = new Map(states.map((state) => [state.key, state]));
  const normalized = definitions.map((definition) =>
    normalizeSingleState(definition, stateMap.get(definition.key), mergedOptions),
  );

  return sortStatesForRender(definitions, normalized).map((state, index) => ({
    ...state,
    order: index,
  }));
}

export function buildAdminTableColumns<RecordType>(
  definitions: AdminTableColumnDefinition<RecordType>[],
  states: AdminTableColumnState[],
) {
  const definitionsMap = toDefinitionsMap(definitions);
  const columns: TableColumnType<RecordType>[] = [];

  sortStatesForRender(definitions, states)
    .filter((state) => state.visible)
    .forEach((state) => {
      const definition = definitionsMap.get(state.key);
      if (!definition) {
        return;
      }

      columns.push({
        ...definition.column,
        fixed: state.fixed || undefined,
        key: definition.key,
        resizable: true,
        width: state.width,
      });
    });

  return columns;
}

export function buildAdminTableColumnSettingsItems<RecordType>(
  definitions: AdminTableColumnDefinition<RecordType>[],
  states: AdminTableColumnState[],
) {
  const definitionsMap = toDefinitionsMap(definitions);
  const orderedStates = sortStatesForRender(definitions, states);
  const visibleBusinessColumnCount = orderedStates.filter((state) => {
    const definition = definitionsMap.get(state.key);
    return state.visible && !definition?.system;
  }).length;

  return orderedStates.map((state): AdminTableColumnSettingsItem => {
    const definition = definitionsMap.get(state.key);
    const isSystem = definition?.system ?? false;

    return {
      ...state,
      disableFixed: isSystem,
      disableVisibility:
        isSystem ||
        (state.visible && !isSystem && visibleBusinessColumnCount <= 1),
      draggable: !isSystem,
      system: isSystem,
      title: definition?.title ?? state.key,
    };
  });
}

export function calculateAdminTableScrollX(
  states: AdminTableColumnState[],
  options: AdminTableColumnOptions = {},
) {
  const mergedOptions = resolveAdminTableColumnOptions(options);
  return states
    .filter((state) => state.visible)
    .reduce((total, state) => total + state.width, mergedOptions.scrollBuffer);
}

export function reorderAdminTableColumnStates<RecordType>(
  definitions: AdminTableColumnDefinition<RecordType>[],
  states: AdminTableColumnState[],
  oldIndex: number,
  newIndex: number,
) {
  if (
    oldIndex === newIndex ||
    oldIndex < 0 ||
    newIndex < 0 ||
    oldIndex >= states.length ||
    newIndex >= states.length
  ) {
    return states;
  }

  const orderedStates = sortStatesForRender(definitions, states);
  const nextStates = [...orderedStates];
  const [movedState] = nextStates.splice(oldIndex, 1);
  if (!movedState) {
    return states;
  }
  nextStates.splice(newIndex, 0, movedState);

  return nextStates.map((state, index) => ({
    ...state,
    order: index,
  }));
}

export function resolveAdminTableColumnKey<RecordType>(
  column: Pick<TableColumnType<RecordType>, 'dataIndex' | 'key'>,
  fallback = '',
) {
  if (column.key != null) {
    return String(column.key);
  }
  if (typeof column.dataIndex === 'number' || typeof column.dataIndex === 'string') {
    return String(column.dataIndex);
  }
  return fallback;
}

export function resolveAdminTableColumnOptions(
  options: AdminTableColumnOptions = {},
): Required<AdminTableColumnOptions> {
  return {
    defaultWidth: options.defaultWidth ?? DEFAULT_WIDTH,
    maxWidth: options.maxWidth ?? MAX_WIDTH,
    minWidth: options.minWidth ?? MIN_WIDTH,
    scrollBuffer: options.scrollBuffer ?? SCROLL_BUFFER,
    systemColumnKeys: options.systemColumnKeys ?? [],
  };
}
