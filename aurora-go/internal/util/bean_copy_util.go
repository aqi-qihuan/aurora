package util

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

// CopyProperties 对象属性拷贝（对标 Java 的 BeanUtil.copyProperties / BeanUtils.copy）
// 将 src 中同名字段值拷贝到 dst（基于JSON中间转换，支持嵌套结构体）
// 注意: 性能不如直接赋值，但通用性更强
func CopyProperties(dst, src interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// CopyPropertiesSelective 选择性拷贝（仅拷贝指定字段）
func CopyPropertiesSelective(dst, src interface{}, fields ...string) error {
	srcMap := structToMap(src)
	dstMap := make(map[string]interface{})

	fieldSet := make(map[string]bool)
	for _, f := range fields {
		fieldSet[strings.ToLower(f)] = true
	}

	for k, v := range srcMap {
		if fieldSet[strings.ToLower(k)] {
			dstMap[k] = v
		}
	}

	return mapToStruct(dst, dstMap)
}

// StructToMap 结构体转 Map（用于灵活的数据组装场景）
func StructToMap(obj interface{}) map[string]interface{} {
	return structToMap(obj)
}

// MapToStruct Map转结构体
func MapToStruct(obj interface{}, m map[string]interface{}) error {
	return mapToStruct(obj, m)
}

// ==================== 内部方法 ====================

func structToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return result
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
	 fieldValue := v.Field(i)

	 // 跳过非导出字段和零值字段
	 if field.PkgPath != "" || fieldValue.IsZero() {
		 continue
	 }

	 // 获取json tag作为key
		tag := field.Tag.Get("json")
		key := strings.Split(tag, ",")[0]
		if key == "" || key == "-" {
			key = field.Name
		}

	 // 处理时间类型特殊序列化
	 switch fv := fieldValue.Interface().(type) {
		 case time.Time:
			 result[key] = fv.Format(time.RFC3339)
		 default:
			 result[key] = fv
		 }
	 }
	return result
}

func mapToStruct(obj interface{}, m map[string]interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
}
