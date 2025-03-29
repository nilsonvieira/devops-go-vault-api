package vault

func SearchAndReplacePassword(basePath, oldPassword, newPassword string) ([]PasswordUpdateResult, error) {
	return SearchAndReplacePasswordDirect(basePath, oldPassword, newPassword, EditMode)
}
