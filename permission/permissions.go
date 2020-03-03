package permission


type PermissionChecker interface {
	CheckPermission(context map[string]interface{}) bool
}


type StandardPermissionChecker struct {
	PermissionName string
	UserId         uint
}

func (c *StandardPermissionChecker) CheckPermission(context map[string]interface{}) bool {
	err, hasPermission := CheckUserHasPermission(c.UserId, CreateBookPermissionName)
	if err != nil {
		return false
	}
	return hasPermission
}
