package datastore

type IDatastoreConfiguration interface {
	Redis() IRedisConfiguration
}

const (
	datastoreKey = "datastore."
)

type DatastoreConfiguration struct{}

func NewDatastoreConfiguration() *DatastoreConfiguration {
	return &DatastoreConfiguration{}
}

func (d *DatastoreConfiguration) Redis() IRedisConfiguration {
	return NewRedisConfiguration()
}
