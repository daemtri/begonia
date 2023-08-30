package types

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/daemtri/begonia/di/box/flagvar"
)

var (
	quotationMark = []byte{'"'}
)

type List[T flagvar.BaseType] struct {
	Items []T
}

func (l List[T]) MarshalText() ([]byte, error) {
	strArr := flagvar.ToStringSlice(l.Items)
	data := strings.Join(strArr, ",")
	return []byte(data), nil
}

func (l List[T]) MarshalJSON() ([]byte, error) {
	strArr := flagvar.ToStringSlice(l.Items)
	data := strings.Join(strArr, ",")
	return []byte(`"` + data + `"`), nil
}

func (l *List[T]) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	if strings.EqualFold(string(text), "null") {
		return nil
	}

	ret := strings.Split(string(text), ",")
	val, err := flagvar.StringSliceTo[T](ret)
	if err != nil {
		return err
	}
	l.Items = val
	return nil
}

func (l *List[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("data长度为0")
	}
	if strings.EqualFold(string(data), "null") {
		return nil
	}
	if data[0] != quotationMark[0] {
		// return fmt.Errorf("%s不是字符串", data)
	} else {
		data = bytes.TrimSuffix(bytes.TrimPrefix(data, quotationMark), quotationMark)
	}
	if len(data) == 0 {
		return nil
	}

	ret := strings.Split(string(data), ",")
	val, err := flagvar.StringSliceTo[T](ret)
	if err != nil {
		return err
	}
	l.Items = val
	return nil
}

// Contains 是否包含
func (l *List[T]) Contains(data T) bool {
	if len(l.Items) == 0 {
		return false
	}
	for _, v := range l.Items {
		if v == data {
			return true
		}
	}
	return false
}

func IndexOf[T flagvar.BaseType](arr []T, index int) T {
	if len(arr)-1 < index {
		var def T
		return def
	}
	return arr[index]
}

// MappingList 把两个切片映射为一个map，第一个切片的值作为key，第二个切片的值作为value
// 如果valueArr长度不及KeyArr，缺少的部分，将会使用类型默认值
// 如果valueArr长度大于keyArr，超过的部分，将会被忽略
func MappingList[K, V flagvar.BaseType](keyArr List[K], valueArr List[V]) map[K]V {
	var def V
	kv := make(map[K]V, len(keyArr.Items))
	max := len(valueArr.Items) - 1
	for i, key := range keyArr.Items {
		if i > max {
			kv[key] = def
		} else {
			kv[key] = valueArr.Items[i]
		}
	}
	return kv
}

func MappingFilter[K flagvar.BaseType, V any](keyArr List[K], valueArr []V, getK func(*V) K) map[K]V {
	var def V
	valueMap := SliceToMap(valueArr, getK)
	kv := make(map[K]V, len(keyArr.Items))
	max := len(valueArr) - 1
	for i, key := range keyArr.Items {
		if i > max {
			kv[key] = def
		} else {
			kv[key] = valueMap[key]
		}
	}
	return kv
}

func SliceToMap[K flagvar.BaseType, V any](arr []V, getKey func(*V) K) map[K]V {
	kv := make(map[K]V, len(arr))
	for i := range arr {
		key := getKey(&arr[i])
		kv[key] = arr[i]
	}
	return kv
}

// SliceFold 把一个Slice分组折叠
func SliceFold[K flagvar.BaseType, V any](arr []V, getKey func(*V) K) map[K][]V {
	kv := make(map[K][]V, len(arr))
	for i := range arr {
		key := getKey(&arr[i])
		_, ok := kv[key]
		if !ok {
			kv[key] = []V{arr[i]}
		} else {
			kv[key] = append(kv[key], arr[i])
		}
	}
	return kv
}
