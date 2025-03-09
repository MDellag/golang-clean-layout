package constants

type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleCustomer UserRole = "customer"
	UserRoleStaff    UserRole = "staff"
)
