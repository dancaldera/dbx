package models

// SupportedDatabaseTypes represents the supported database types in DBX
var SupportedDatabaseTypes = []DBType{
	{Name: "PostgreSQL", Driver: "postgres"},
	{Name: "MySQL", Driver: "mysql"},
	{Name: "SQLite", Driver: "sqlite3"},
}
