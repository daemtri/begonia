package helper

import (
	"reflect"
)

func reflectType[T any]() reflect.Type {
	return reflect.TypeOf(new(T)).Elem()
}

func zeroValue[T any]() (x T) { return }

func Zero[T any]() T {
	typ := reflectType[T]()
	switch typ.Kind() {
	case reflect.Interface:
		return zeroValue[T]()
	case reflect.Pointer:
		return reflect.New(typ.Elem()).Interface().(T)
	default:
		return reflect.New(typ).Elem().Interface().(T)
	}
}

func ZeroWithKind[T any]() (T, reflect.Kind) {
	typ := reflectType[T]()
	switch typ.Kind() {
	case reflect.Interface:
		return zeroValue[T](), reflect.Interface
	case reflect.Pointer:
		return reflect.New(typ.Elem()).Interface().(T), reflect.Pointer
	default:
		return reflect.New(typ).Elem().Interface().(T), typ.Kind()
	}
}
