package binding

import (
	"sync/atomic"

	"fyne.io/fyne/v2"
)

func newBaseItem[T any](comparator func(T, T) bool) *baseItem[T] {
	return &baseItem[T]{val: new(T), comparator: comparator}
}

func newBaseItemComparable[T bool | float64 | int | rune | string]() *baseItem[T] {
	return newBaseItem[T](func(a, b T) bool { return a == b })
}

type baseItem[T any] struct {
	base

	comparator func(T, T) bool
	val        *T
}

func (b *baseItem[T]) Get() (T, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.val == nil {
		return *new(T), nil
	}
	return *b.val, nil
}

func (b *baseItem[T]) Set(val T) error {
	b.lock.Lock()
	equal := b.comparator(*b.val, val)
	*b.val = val
	b.lock.Unlock()

	if !equal {
		b.trigger()
	}

	return nil
}

func baseBindExternal[T any](val *T, comparator func(T, T) bool) *baseExternalItem[T] {
	if val == nil {
		val = new(T) // never allow a nil value pointer
	}
	b := &baseExternalItem[T]{}
	b.comparator = comparator
	b.val = val
	b.old = *val
	return b
}

func baseBindExternalComparable[T bool | float64 | int | rune | string](val *T) *baseExternalItem[T] {
	if val == nil {
		val = new(T) // never allow a nil value pointer
	}
	b := &baseExternalItem[T]{}
	b.comparator = func(a, b T) bool { return a == b }
	b.val = val
	b.old = *val
	return b
}

type baseExternalItem[T any] struct {
	baseItem[T]

	old T
}

func (b *baseExternalItem[T]) Set(val T) error {
	b.lock.Lock()
	if b.comparator(b.old, val) {
		b.lock.Unlock()
		return nil
	}
	*b.val = val
	b.old = val
	b.lock.Unlock()

	b.trigger()
	return nil
}

func (b *baseExternalItem[T]) Reload() error {
	return b.Set(*b.val)
}

type prefBoundBase[T bool | float64 | int | string] struct {
	base
	key   string
	get   func(string) T
	set   func(string, T)
	cache atomic.Pointer[T]
}

func (b *prefBoundBase[T]) Get() (T, error) {
	cache := b.get(b.key)
	b.cache.Store(&cache)
	return cache, nil
}

func (b *prefBoundBase[T]) Set(v T) error {
	b.set(b.key, v)

	b.lock.RLock()
	defer b.lock.RUnlock()
	b.trigger()
	return nil
}

func (b *prefBoundBase[T]) setKey(key string) {
	b.key = key
}

func (b *prefBoundBase[T]) checkForChange() {
	val := b.cache.Load()
	if val != nil && b.get(b.key) == *val {
		return
	}
	b.trigger()
}

type genericItem[T any] interface {
	DataItem
	Get() (T, error)
	Set(T) error
}

func lookupExistingBinding[T any](key string, p fyne.Preferences) (genericItem[T], bool) {
	binds := prefBinds.getBindings(p)
	if binds == nil {
		return nil, false
	}

	if listen, ok := binds.Load(key); listen != nil && ok {
		if l, ok := listen.(genericItem[T]); ok {
			return l, ok
		}
		fyne.LogError(keyTypeMismatchError+key, nil)
	}

	return nil, false
}