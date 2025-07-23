package rest

type Record []string

type DB interface {
	Create(r Record) error
	Update(r Record) error
	Get(id string) (Record, error)
	Delete(id string) error
	Iter() func(yield func(Record, error) bool)
	Close() error
}
