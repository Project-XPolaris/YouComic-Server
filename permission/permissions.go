package permission


type PermissionChecker interface {
	CheckPermission() bool
}


type StandardPermissionChecker struct {
	PermissionName string
	UserId         uint
}

func (c *StandardPermissionChecker) CheckPermission() bool {
	err, hasPermission := CheckUserHasPermission(c.UserId, CreateBookPermissionName)
	if err != nil {
		return false
	}
	return hasPermission
}


