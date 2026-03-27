package treeUtils

import "reflect"

// TreeFieldConfig 定义树构建所需的字段配置
type TreeFieldConfig[T any] struct {
	GetID       func(T) interface{} // 获取节点ID的函数
	GetParentID func(T) interface{} // 获取父节点ID的函数
	SetChildren func(T, []T) T      // 设置子节点的函数
}

// TreeBuilder 泛型树构建器
type TreeBuilder[T any] struct {
	config TreeFieldConfig[T]
}

// NewTreeBuilder 创建新的树构建器实例
func NewTreeBuilder[T any](config TreeFieldConfig[T]) *TreeBuilder[T] {
	return &TreeBuilder[T]{config: config}
}

// BuildTree 构建树形结构，保持原始结构体类型
func (tb *TreeBuilder[T]) BuildTree(data []T) []T {
	if len(data) == 0 {
		return nil
	}
	// 构建父子关系映射
	childrenMap := make(map[interface{}][]T)
	var rootNodes []T
	for _, item := range data {
		parentID := tb.config.GetParentID(item)
		if tb.isNilOrEmpty(parentID) {
			rootNodes = append(rootNodes, item)
		} else {
			childrenMap[parentID] = append(childrenMap[parentID], item)
		}
	}
	// 递归设置子节点
	return tb.processNodes(rootNodes, childrenMap)
}

// processNodes 处理节点并设置子节点
func (tb *TreeBuilder[T]) processNodes(nodes []T, childrenMap map[interface{}][]T) []T {
	result := make([]T, len(nodes))
	for i, node := range nodes {
		id := tb.config.GetID(node)
		if children, exists := childrenMap[id]; exists {
			// 递归处理子节点
			processedChildren := tb.processNodes(children, childrenMap)
			// 使用配置的函数设置子节点
			result[i] = tb.config.SetChildren(node, processedChildren)
		} else {
			// 没有子节点时也调用设置函数（通常用于初始化空切片）
			result[i] = tb.config.SetChildren(node, []T{})
		}
	}
	return result
}

// isNilOrEmpty 判断值是否为空
func (tb *TreeBuilder[T]) isNilOrEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	default:
		return false
	}
}
