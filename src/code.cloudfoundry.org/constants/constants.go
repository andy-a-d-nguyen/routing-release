package constants

type DatabaseDriver string

const (
	Postgres               DatabaseDriver = "postgres"
	PostgresUserName                      = "postgres"
	PostgresPassword                      = ""
	MySQL                  DatabaseDriver = "mysql"
	MySQLUsername                         = "root"
	MYSQLPassword                         = "password"
	RoutingAPIIP                          = "127.0.0.1"
	RoutingAPIHost                        = "localhost"
	RoutingAPISystemDomain                = "example.com"
)
