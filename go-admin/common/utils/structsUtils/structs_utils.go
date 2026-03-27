package structsUtils

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"strconv"
	"sync"
	"time"
)

// StructFieldValues 提取结构体切片中的某一个字段值组成切片
func StructFieldValues[T any, K any](list *[]T, fieldName string) ([]K, error) {
	var result []K
	if len(*list) == 0 {
		return result, nil
	}

	elemType := reflect.TypeOf((*list)[0])
	if _, ok := elemType.FieldByName(fieldName); !ok {
		return nil, fmt.Errorf("field %s not found in struct", fieldName)
	}

	fieldIndex := -1
	for i := 0; i < elemType.NumField(); i++ {
		if elemType.Field(i).Name == fieldName {
			fieldIndex = i
			break
		}
	}

	if fieldIndex == -1 {
		return nil, fmt.Errorf("field %s not found", fieldName)
	}

	for _, item := range *list {
		val := reflect.ValueOf(item)
		fieldVal := val.Field(fieldIndex)

		// 尝试将字段值转换为 K 类型
		kValPtr := (*K)(nil) // 获取 K 的类型信息
		kType := reflect.TypeOf(kValPtr).Elem()

		if fieldVal.Type() != kType {
			return nil, fmt.Errorf("field %s type %v does not match expected type %v", fieldName, fieldVal.Type(), kType)
		}

		result = append(result, fieldVal.Interface().(K))
	}

	return result, nil
}

// CopyBeanProp 将 src 结构体字段的值复制到 dest 中同名字段上
func CopyBeanProp(dest, src interface{}) {
	if src == nil {
		return
	}

	srcVal := reflect.ValueOf(src)
	srcType := reflect.TypeOf(src)
	if srcVal.Kind() == reflect.Ptr {
		srcVal = reflect.ValueOf(src).Elem()
		srcType = srcVal.Type()
	}

	destVal := reflect.ValueOf(dest).Elem() // 解引用指针
	destType := destVal.Type()

	for i := 0; i < srcType.NumField(); i++ {
		srcField := srcType.Field(i)
		destField, ok := destType.FieldByName(srcField.Name)
		if !ok {
			continue // 跳过目标中不存在的字段
		}
		// 获取 src 字段的值
		srcFieldValue := srcVal.Field(i)
		// 类型必须一致才能赋值
		destFieldValue := destVal.FieldByName(srcField.Name)
		if destField.Type != srcField.Type {
			destValue, isOk := ConvertValue(srcFieldValue, destField.Type)
			if !isOk {
				fmt.Sprintln("src:%s TO target:%s FROM %s TO %s Failed", srcField.Name, destField.Name, srcField.Type, destField.Type)
				continue
			}
			if destFieldValue.CanSet() {
				destFieldValue.Set(destValue)
			}
		} else {
			if destFieldValue.CanSet() {
				destFieldValue.Set(srcFieldValue)
			}
		}
	}
}

// ListGroupByFunc 根据 函数返回值分组
func ListGroupByFunc[T any, K comparable](list *[]T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T)

	for _, item := range *list {
		key := keyFunc(item)
		result[key] = append(result[key], item)
	}

	return result
}

func ListToMapByFunc[T any, K comparable](list *[]T, keyFunc func(T) K) map[K]T {
	result := make(map[K]T)

	for _, item := range *list {
		key := keyFunc(item)
		result[key] = item
	}

	return result
}

// UniqueKey 结构体数组去重 结构体要先实现 UniqueKey 接口
type UniqueKey interface {
	UniqueKey() any
}

func DeduplicateByUniqueKey[T UniqueKey](list *[]T) []T {
	seen := make(map[any]bool)
	result := make([]T, 0)

	for _, item := range *list {
		key := item.UniqueKey()
		if _, exists := seen[key]; !exists {
			seen[key] = true
			result = append(result, item)
		}
	}

	return result
}

type OrderIDGenerator struct {
	lastNano   int64
	counter    uint
	counterMax uint
	mu         sync.Mutex
}

func NewOrderIDGenerator() *OrderIDGenerator {
	return &OrderIDGenerator{
		lastNano:   0,
		counter:    0,
		counterMax: 99999,
	}
}

// Generate20BitID 生成20位纯数字ID
func (g *OrderIDGenerator) Generate20BitID() int64 {
	now := time.Now()
	// 纳秒级时间戳用于高精度区分
	nowNano := now.UnixNano()

	g.mu.Lock()
	defer g.mu.Unlock()

	if nowNano > g.lastNano {
		g.counter = 0
	} else {
		g.counter++
		if g.counter > g.counterMax {
			// 超出当前时间精度的最大数量，等待到下一毫秒
			for nowNano <= g.lastNano {
				time.Sleep(time.Nanosecond)
				now = time.Now()
				nowNano = now.UnixNano()
			}
			g.counter = 0
		}
	}
	g.lastNano = nowNano

	// 提取时间部分
	year := now.Year() % 100
	month := now.Month()
	day := now.Day()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()
	millisecond := now.Nanosecond() / 1e6 // 微秒转毫秒

	// 构造时间部分（12位）
	timePart := fmt.Sprintf("%02d%02d%02d%02d%02d%02d",
		year, month, day, hour, minute, second)

	// 拼接毫秒+序列号（共8位）
	if g.counter == 0 {
		// 生成 4 位随机数：1000 ~ 9999
		g.counter = uint(rand.IntN(9000)) + 1000
	}
	seqPart := fmt.Sprintf("%03d%04d", millisecond, g.counter)
	numstr := timePart + seqPart
	// 返回完整ID（19位）
	uqid, _ := strconv.ParseInt(numstr, 10, 64)
	return uqid
}

func ConvertValue(val reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
	srcKind := val.Kind()
	targetKind := targetType.Kind()

	// 如果是接口或指针，先解引用（简化处理）
	for val.Kind() == reflect.Interface {
		if val.IsNil() {
			return reflect.Zero(targetType), false
		}
		val = val.Elem()
	}

	switch targetKind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var x int64
		switch srcKind {
		case reflect.String:
			s := val.String()
			if s == "" {
				x = 0
			} else {
				var err error
				x, err = strconv.ParseInt(s, 10, 64)
				if err != nil {
					return reflect.Zero(targetType), false
				}
			}
		case reflect.Float32, reflect.Float64:
			x = int64(val.Float())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x = int64(val.Uint())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			x = val.Int()
		default:
			return reflect.Zero(targetType), false
		}
		return reflect.ValueOf(x).Convert(targetType), true

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var x uint64
		switch srcKind {
		case reflect.String:
			s := val.String()
			if s == "" {
				x = 0
			} else {
				var err error
				x, err = strconv.ParseUint(s, 10, 64)
				if err != nil {
					return reflect.Zero(targetType), false
				}
			}
		case reflect.Float32, reflect.Float64:
			x = uint64(val.Float())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			x = uint64(val.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x = val.Uint()
		default:
			return reflect.Zero(targetType), false
		}
		return reflect.ValueOf(x).Convert(targetType), true

	case reflect.Float32, reflect.Float64:
		var x float64
		switch srcKind {
		case reflect.String:
			s := val.String()
			if s == "" {
				x = 0
			} else {
				var err error
				x, err = strconv.ParseFloat(s, 64)
				if err != nil {
					return reflect.Zero(targetType), false
				}
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			x = float64(val.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x = float64(val.Uint())
		case reflect.Float32, reflect.Float64:
			x = val.Float()
		default:
			return reflect.Zero(targetType), false
		}
		return reflect.ValueOf(x).Convert(targetType), true

	case reflect.String:
		switch srcKind {
		case reflect.String:
			return val.Convert(targetType), true
		default:
			return reflect.ValueOf(fmt.Sprintf("%v", val.Interface())).Convert(targetType), true
		}
	default:
		return reflect.Zero(targetType), false
	}
}

// HasField 判断结构体或指针是否含有指定字段
func HasField(v interface{}, field string) bool {
	rv := reflect.ValueOf(v)

	// 如果是指针，获取其指向的元素
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// 必须是结构体
	if rv.Kind() != reflect.Struct {
		return false
	}

	// 获取结构体的类型信息
	st := rv.Type()

	// 查找字段（区分大小写）
	_, found := st.FieldByName(field)
	return found
}

// StructToMap 将结构体转换为 map[string]interface{}，递归处理嵌套结构体和切片
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	v := reflect.ValueOf(obj)

	// 如果是指针，解引用
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("input is nil pointer")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct or pointer to struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 跳过未导出字段
		if !fieldType.IsExported() {
			continue
		}

		// 获取 json 标签
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// 解析标签（如 "name,omitempty"）
		tagParts := SplitJsonTag(jsonTag)
		key := tagParts[0]

		// 递归处理字段值
		value, err := convertField(field)
		if err != nil {
			return nil, fmt.Errorf("error converting field %s: %v", fieldType.Name, err)
		}

		result[key] = value
	}

	return result, nil
}

// convertField 递归处理字段值
func convertField(v reflect.Value) (interface{}, error) {
	// 如果是指针，解引用
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		return convertField(v.Elem())
	}

	// 如果是结构体，递归转为 map
	if v.Kind() == reflect.Struct {
		return StructToMap(v.Interface())
	}

	// 如果是切片，逐个处理元素
	if v.Kind() == reflect.Slice {
		slice := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			elem, err := convertField(v.Index(i))
			if err != nil {
				return nil, err
			}
			slice[i] = elem
		}
		return slice, nil
	}

	// 基础类型（int, string, bool, float 等）直接返回
	return v.Interface(), nil
}

// SplitJsonTag 简单解析 json 标签，提取 key 名称
func SplitJsonTag(tag string) []string {
	if tag == "" {
		return nil
	}
	// 找第一个逗号
	for i, ch := range tag {
		if ch == ',' {
			return []string{tag[:i], tag[i+1:]}
		}
	}
	return []string{tag}
}
