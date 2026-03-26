import type { SelectProps } from 'ant-design-vue';

export interface AdminFormFieldOption {
  disabled?: boolean;
  label: string;
  value: number | string;
}

export interface AdminFormFieldSchema {
  allowClear?: boolean;
  component:
    | 'input'
    | 'input-number'
    | 'select'
    | 'textarea'
    | 'tree-select';
  disabled?: boolean;
  field: string;
  filterOption?: SelectProps['filterOption'];
  hidden?: boolean;
  itemClass?: string;
  label: string;
  loading?: boolean;
  min?: number;
  options?: AdminFormFieldOption[];
  placeholder?: string;
  props?: Record<string, unknown>;
  required?: boolean;
  rows?: number;
  showSearch?: boolean;
  span?: 1 | 2;
}
