package treeUtils

import "reflect"

// SimpleTreeBuilder 简单树构建器，通过字段名配置
type SimpleTreeBuilder[T any] struct {
	idField       string // ID字段名
	parentIDField string // 父ID字段名
	childrenField string // 子节点集合字段名
}

// NewSimpleTreeBuilder 创建简单树构建器
func NewSimpleTreeBuilder[T any](idField string, parentIDField string, childrenField string) *SimpleTreeBuilder[T] {
	return &SimpleTreeBuilder[T]{
		idField:       idField,
		parentIDField: parentIDField,
		childrenField: childrenField,
	}
}

// BuildTree 使用字段名构建树形结构
func (stb *SimpleTreeBuilder[T]) BuildTree(data []T) []T {
	if len(data) == 0 {
		return nil
	}
	// 构建父子关系映射
	childrenMap := make(map[interface{}][]T)
	var rootNodes []T
	for _, item := range data {
		parentID := stb.getFieldValue(item, stb.parentIDField)
		if stb.isNilOrEmpty(parentID) {
			rootNodes = append(rootNodes, item)
		} else {
			childrenMap[parentID] = append(childrenMap[parentID], item)
		}
	}
	// 递归设置子节点
	return stb.processNodes(rootNodes, childrenMap)
}

// getFieldValue 通过字段名获取字段值
func (stb *SimpleTreeBuilder[T]) getFieldValue(obj T, fieldName string) interface{} {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		panic("字段 " + fieldName + " 不存在")
	}
	return field.Interface()
}

// setChildrenField 通过字段名设置子节点
func (stb *SimpleTreeBuilder[T]) setChildrenField(obj T, children []T) T {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// 创建可修改的副本
	newV := reflect.New(v.Type()).Elem()
	newV.Set(v)
	childrenField := newV.FieldByName(stb.childrenField)
	if !childrenField.IsValid() || !childrenField.CanSet() {
		panic("子节点字段 " + stb.childrenField + " 不存在或不可设置")
	}
	if childrenField.Kind() != reflect.Slice {
		panic("子节点字段 " + stb.childrenField + " 不是切片类型")
	}
	childrenValues := reflect.MakeSlice(childrenField.Type(), len(children), len(children))
	for i, child := range children {
		childrenValues.Index(i).Set(reflect.ValueOf(child))
	}
	childrenField.Set(childrenValues)
	return newV.Interface().(T)
}

// processNodes 处理节点并设置子节点
func (stb *SimpleTreeBuilder[T]) processNodes(nodes []T, childrenMap map[interface{}][]T) []T {
	result := make([]T, len(nodes))
	for i, node := range nodes {
		id := stb.getFieldValue(node, stb.idField)
		if children, exists := childrenMap[id]; exists {
			processedChildren := stb.processNodes(children, childrenMap)
			result[i] = stb.setChildrenField(node, processedChildren)
		} else {
			result[i] = stb.setChildrenField(node, []T{})
		}
	}
	return result
}

// isNilOrEmpty 判断值是否为空
func (stb *SimpleTreeBuilder[T]) isNilOrEmpty(value interface{}) bool {
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
