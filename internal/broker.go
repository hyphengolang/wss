package internal

// Material
//
// https://www.cs.cmu.edu/~410-s05/lectures/L31_LockFree.pdf
type Broker[K, V any] interface {
	Load(key K) (value V, ok bool)
	LoadOrStore(key K, value V) (actual V, loaded bool)
	Store(key K, value V)
	Delete(key K)
}
