package helper

import (
	"reflect"
)

func reflectType[T any]() reflect.Type {
	return reflect.TypeOf(new(T)).Elem()
}

func emptyValue[T any]() (x T) { return }

func New[T any]() T {
	typ := reflectType[T]()
	switch typ.Kind() {
	case reflect.Interface:
		return emptyValue[T]()
	case reflect.Pointer:
		return reflect.New(typ.Elem()).Interface().(T)
	default:
		return reflect.New(typ).Elem().Interface().(T)
	}
}

func NewWithKind[T any]() (T, reflect.Kind) {
	typ := reflectType[T]()
	switch typ.Kind() {
	case reflect.Interface:
		return emptyValue[T](), reflect.Interface
	case reflect.Pointer:
		return reflect.New(typ.Elem()).Interface().(T), reflect.Pointer
	default:
		return reflect.New(typ).Elem().Interface().(T), typ.Kind()
	}
}
