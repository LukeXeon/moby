package osl

// neigh holds the options of a neighbor entry; only the overlay driver, which
// android does not register, adds neighbors.
type neigh struct {
	linkName string
	family   int
}
