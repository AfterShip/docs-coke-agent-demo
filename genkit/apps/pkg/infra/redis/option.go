package redis

type Options struct {
	MasterName   string   `json:"master_name" mapstructure:"master_name" validate:"required"`
	SentinelAddr []string `json:"sentinel_addr" mapstructure:"sentinel_addr" validate:"required"`
	Password     string   `json:"password" mapstructure:"password"`
	DBNumber     int      `json:"db_number" mapstructure:"db_number"`
	PoolSize     int      `json:"pool_size" mapstructure:"pool_size"`
}

func NewRedisOption() *Options {
	return &Options{
		PoolSize:     120,
		DBNumber:     1,
		MasterName:   "mymaster",
		SentinelAddr: []string{},
		Password:     "",
	}
}
