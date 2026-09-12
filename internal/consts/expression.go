package consts

type expression string

// ~
const (
	ExprNull         expression = "$NULL"
	ExprBlank        expression = "$BLANK"
	ExprIndetermined expression = "$INDETERMINED"

	ExprUnreachableCase    expression = "unreachable case contacted"
	ExprUnsupportedFeature expression = "feature unsupported"

	ExprFailedRegistercodeWithdraw expression = "failed to withdraw registercode"

	ExprBadFormat expression = "bad format"

	//ExprAtomid       expression = "Attr-AtomID"
	ExprUsername     expression = "Attr-Username"
	ExprNickname     expression = "Attr-Nickname"
	ExprPassword     expression = "Attr-Password"
	ExprEntrycode    expression = "Attr-Entrycode"
	ExprRegistercode expression = "Attr-Registercode"

	InlineExprUser  expression = "User"
	InlineExprAdmin expression = "Admin"
	ExprRoleUser    expression = "Role-User"
	ExprRoleAdmin   expression = "Role-Admin"

	ExprCalendar expression = "Attr-Calendar"

	ExprEnckey expression = "Crypto-PrivateKey"
	ExprDeckey expression = "Crypto-PublicKey"
)

// db
const (
	ExprRegistercodeRecord expression = "Record-Registercode"
)

// http response text
const (
	HttpExprInternalServerError expression = "internal server error"
	HttpExprOk                  expression = "ok"
	HttpExprTooManyRequests     expression = "too many requests"
	HttpExprErrorUnknown        expression = "unknown error"
	HttpExprErrorBadRequest     expression = "bad request"
)

// program
const (
	CtxExprKeyJWT    expression = "jwt_user"
	CtxExprKeyDomain expression = "router_domain"
	ExprUserID       expression = "uid"
	ExprTraceid      expression = "traceid"
)

// json
// only expressions in the json body can be place here.
const (
	JsonExprAtomid       expression = "atomid"
	JsonExprUsername     expression = "username"
	JsonExprNickname     expression = "nickname"
	JsonExprPassword     expression = "password"
	JsonExprRegistercode expression = "registercode"
	JsonExprEntrycode    expression = "entrycode"
	JsonExprAccessToken  expression = "access_token"
	JsonExprRefreshToken expression = "refresh_token"
	JsonExprRole         expression = "role"
	JsonExprTraceid      expression = "traceid"
	JsonExprCallChain    expression = "call_chain"

	JsonExprCalendarID expression = "calendar_id"
	JsonExprRecords    expression = "records"
	JsonExprWeekday    expression = "weekday"
	JsonExprStartMin   expression = "start_min"
	JsonExprEndMin     expression = "end_min"
	JsonExprTitle      expression = "title"
)

// process
const (
	ExprRegister expression = "Process-Register"
	ExprLogin    expression = "Process-Login"
)

// module
const (
	ModExprMain expression = "main"

	ModExprHandler         expression = "handler"
	ModExprHandlerRegister expression = "Handler-Register"
	ModExprHandlerLogin    expression = "Handler-Login"
	ModExprHandlerAdmin    expression = "Handler-Admin"
	ModExprHandlerMe       expression = "Handler-Me"
	ModExprHandlerRefresh  expression = "Handler-Refresh"
	ModExprHandlerSignout  expression = "Handler-Signout"
	ModExprHandlerCalendar expression = "Handler-Calendar"

	ModExprService      expression = "service"
	ModExprServiceUser  expression = "Service-User"
	ModExprServiceToken expression = "Service-Token"
	ModExprServiceAdmin expression = "Service-Admin"

	ModExprMiddlware            expression = "middleware"
	ModExprMiddlwareAuthen      expression = "Middleware-Authen"
	ModExprMiddlwareInjector    expression = "Middleware-Injector"
	ModExprMiddlwareRateLimiter expression = "Middleware-RateLimiter"
	ModExprMiddlwareRoleControl expression = "Middleware-RoleControl"

	ModExprUtils expression = "utils"
)

// type of request
const (
	ExprReqRegister  expression = "Request-Register"
	ExprReqLogin     expression = "Request-Login"
	ExprReqUserInfo  expression = "Request-UserInfo"
	ExprReqAdminSudo expression = "Request-AdminSudo"
	ExprReqRefresh   expression = "Request-Refresh"
	ExprReqSignout   expression = "Request-Signout"

	ExprReqCalendarGet    expression = "Request-CalendarGet"
	ExprReqCalendarPut    expression = "Request-CalendarPut"
	ExprReqCalendarDelete expression = "Request-CalendarDelete"
)

// router domain
const (
	DomainExprNull         expression = "Domain-Null"
	DomainExprPrivate      expression = "Domain-Private"
	DomainExprPublic       expression = "Domain-Public"
	DomainExprUnauthorized expression = "Domain-Unauthorized"
)
