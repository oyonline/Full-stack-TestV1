export const TEMPLATE_CATEGORY_OPTIONS = [
  { label: 'CRUD', value: 'crud' },
  { label: '树表', value: 'tree' },
];

export const QUERY_TYPE_OPTIONS = [
  { label: '等于', value: 'EQ' },
  { label: '不等于', value: 'NE' },
  { label: '大于', value: 'GT' },
  { label: '大于等于', value: 'GTE' },
  { label: '小于', value: 'LT' },
  { label: '小于等于', value: 'LTE' },
  { label: '包含', value: 'LIKE' },
  { label: '左模糊', value: 'LEFT_LIKE' },
  { label: '右模糊', value: 'RIGHT_LIKE' },
  { label: '范围', value: 'BETWEEN' },
];

export const HTML_TYPE_OPTIONS = [
  { label: '输入框', value: 'input' },
  { label: '文本域', value: 'textarea' },
  { label: '富文本', value: 'editor' },
  { label: '下拉框', value: 'select' },
  { label: '单选框', value: 'radio' },
  { label: '复选框', value: 'checkbox' },
  { label: '开关', value: 'switch' },
  { label: '日期', value: 'date' },
  { label: '日期时间', value: 'datetime' },
  { label: '图片上传', value: 'imageUpload' },
  { label: '文件上传', value: 'fileUpload' },
];

export const BOOLEAN_STRING_OPTIONS = [
  { label: '是', value: '1' },
  { label: '否', value: '0' },
];

export const DATA_SCOPE_OPTIONS = [
  { label: '关闭', value: 0 },
  { label: '开启', value: 1 },
];

export const ACTIONS_OPTIONS = [
  { label: '不生成', value: 1 },
  { label: '生成', value: 2 },
];

export const AUTH_OPTIONS = [
  { label: '检查权限', value: 1 },
  { label: '跳过权限', value: 2 },
];

export const LOGICAL_DELETE_OPTIONS = [
  { label: '关闭', value: '0' },
  { label: '开启', value: '1' },
];
