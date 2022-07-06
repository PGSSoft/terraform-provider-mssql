package sql

type DatabaseId int

type GenericDatabasePrincipalId int

type UserId GenericDatabasePrincipalId

type DatabaseRoleId GenericDatabasePrincipalId

const EmptyDatabasePrincipalId GenericDatabasePrincipalId = -1

type LoginId string

type DatabasePrincipalId interface {
	UserId | DatabaseRoleId | GenericDatabasePrincipalId
}

type NumericObjectId interface {
	DatabaseId | DatabasePrincipalId
}

type StringObjectId interface {
	LoginId
}

type ObjectId interface {
	NumericObjectId | StringObjectId
}

type DatabasePrincipalType int

const (
	UNONOWN DatabasePrincipalType = iota
	SQL_USER
	DATABASE_ROLE
)
