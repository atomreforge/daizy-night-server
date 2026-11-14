package errs

func BuildErrRegisterLogic(errtype ErrType, http int, field string) *ErrRegisterLogic {
	return &ErrRegisterLogic{Type: errtype, Http: http, Field: field}
}

func BuildErrValidation(errtype ErrType, http int, field string, value string) *ErrValidation {
	return &ErrValidation{Type: errtype, Http: http, Field: field, Value: value}
}

func BuildErrUserLogin(errtype ErrType, http int, user string) *ErrUserLogin {
	return &ErrUserLogin{Type: errtype, Http: http, User: user}
}

func BuildErrRegistercode(errtype ErrType, http int) *ErrRegistercode {
	return &ErrRegistercode{Type: errtype, Http: http}
}

func BuildErrSupport(errtype ErrType, http int) *ErrSupport {
	return &ErrSupport{Type: errtype, Http: http}
}

func BuildErrUnknown(errtype ErrType, http int) *ErrUnknown {
	return &ErrUnknown{Type: errtype, Http: http}
}

func BuildErrDbRecord(errtype ErrType, http int, field string) *ErrDbRecord {
	return &ErrDbRecord{Type: errtype, Http: http, Field: field}
}

func BuildErrJwtToken(errtype ErrType, http int) *ErrJwtToken {
	return &ErrJwtToken{Type: errtype, Http: http}
}

func BuildErrConnDb(errtype ErrType, http int) *ErrConnDb {
	return &ErrConnDb{Type: errtype, Http: http}
}

func BuildErrInternalCall(errtype ErrType) *ErrInternalServiceCall {
	return &ErrInternalServiceCall{Type: errtype}
}

func BuildErrUnnecBiz(errtype ErrType, http int, field string) *ErrUnnecBiz {
	return &ErrUnnecBiz{Type: errtype, Http: http, Field: field}
}
