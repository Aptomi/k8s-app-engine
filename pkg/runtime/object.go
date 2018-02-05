package runtime

// Object represents minimal object that could be operated in runtime with only Kind being mandatory characteristic
type Object interface {
	GetKind() Kind
}

// Storable represents runtime object that could be stored in database and having two additional mandatory characteristics:
// Name and Namespace that together with Kind forms Key (namespace + kind + name) that represents coordinates of the
// object in database
type Storable interface {
	Object
	GetName() string
	GetNamespace() string
}

// Versioned extends Storable with mandatory Generation characteristic to represent versioned objects that are having
// multiple generations stored in database
type Versioned interface {
	Storable
	GetGeneration() Generation
	SetGeneration(gen Generation)
}

// Deletable extends Versioned with deleted marker to represent versioned objects that could be marked as deleted,
// but then created again even with the same content. Best example of such objects is policy objects (and the only for now).
type Deletable interface {
	Versioned
	IsDeleted() bool
	SetDeleted(bool)
}
