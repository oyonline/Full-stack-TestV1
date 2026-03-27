package collectors

import "reflect"

// GroupBy 按照指定键函数对切片进行分组
func GroupBy[T any, K comparable](slice []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range slice {
		key := keyFunc(item)
		result[key] = append(result[key], item)
	}
	return result
}

// ToMapDynamic 支持动态返回任意类型的 map，key 和 value 可根据业务需求动态实现
func ToMapDynamic[T any, K comparable, V any](slice []T, keyFunc func(T) K, valueFunc func(T) V) map[K]V {
	result := make(map[K]V)
	for _, item := range slice {
		key := keyFunc(item)
		value := valueFunc(item)
		result[key] = value
	}
	return result
}

// GroupByAndDistinct 按照动态 key 分组，并返回指定字段的去重集合
func GroupByAndDistinct[T any, K comparable, V comparable](slice []T, keyFunc func(T) K, valueFunc func(T) V) map[K][]V {
	//对每组提取字段值并去重
	result := make(map[K][]V)
	// 用于跟踪每组的去重值
	distinctTracker := make(map[K]map[V]struct{})
	for _, item := range slice {
		key := keyFunc(item)
		value := valueFunc(item)
		// 获取或初始化内层 map
		innerMap, exists := distinctTracker[key]
		if !exists {
			innerMap = make(map[V]struct{})
			distinctTracker[key] = innerMap
		}
		// 去重逻辑
		if _, exists = innerMap[value]; !exists {
			innerMap[value] = struct{}{}
			result[key] = append(result[key], value)
		}
	}
	return result
}

// DistinctField 提取指定字段并去重，返回去重后的字段值切片
func DistinctField[T any, V comparable](slice []T, valueFunc func(T) V) []V {
	result := make([]V, 0)
	seen := make(map[V]struct{})

	for _, item := range slice {
		value := valueFunc(item)
		if _, exists := seen[value]; !exists {
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

// CopyFieldsWithIgnore 复制字段但忽略指定字段
// ignoreFields: 要忽略的字段名列表
func CopyFieldsWithIgnore[Source any, Target any](source *Source, target *Target, ignoreFields ...string) {
	if source == nil || target == nil {
		return
	}

	ignoreMap := make(map[string]bool)
	for _, field := range ignoreFields {
		ignoreMap[field] = true
	}

	sourceValue := reflect.ValueOf(source).Elem()
	targetValue := reflect.ValueOf(target).Elem()

	for i := 0; i < sourceValue.NumField(); i++ {
		sourceField := sourceValue.Type().Field(i)
		fieldName := sourceField.Name

		// 跳过私有字段和忽略的字段
		if (fieldName[0] >= 'a' && fieldName[0] <= 'z') || ignoreMap[fieldName] {
			continue
		}

		targetField := targetValue.FieldByName(fieldName)

		if targetField.IsValid() && targetField.CanSet() && sourceValue.Field(i).Type() == targetField.Type() {
			targetField.Set(sourceValue.Field(i))
		}
	}
}
