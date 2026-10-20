package datastore

//go:generate mockgen -destination=../../mocks/mock_datastore_configuration.go -package=mocks . IDatastoreConfiguration,IRedisConfiguration,IValkeyConfiguration
type IDatastoreConfiguration interface {
	Redis() IRedisConfiguration
	Valkey() IValkeyConfiguration
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

func (d *DatastoreConfiguration) Valkey() IValkeyConfiguration {
	return NewValkeyConfiguration()
}
