package consts

type expression string

// ~
const (
	ExprValueNull    expression = "$NULL"
	ExprValueBlank   expression = "$BLANK"
	ExprIndetermined expression = "$INDETERMINED"

	ExprUnreachableCase    expression = "unreachable case contacted"
	ExprUnsupportedFeature expression = "feature unsupported"

	ExprFailedRegistercodeWithdraw expression = "failed to withdraw registercode"

	ExprBadFormat expression = "bad format"

	//ExprAtomid       expression = "Attr-AtomID"
	UiExprUsername     expression = "Attr-Username"
	UiExprNickname     expression = "Attr-Nickname"
	UiExprPassword     expression = "Attr-Password"
	UiExprEntrycode    expression = "Attr-Entrycode"
	UiExprRegistercode expression = "Attr-Registercode"
	UiExprNid          expression = "Attr-NotificationID"
	UiExprCalendar     expression = "Attr-Calendar"

	InlineExprUser  expression = "User"
	InlineExprAdmin expression = "Admin"
	UiExprRoleUser  expression = "Role-User"
	UiExprRoleAdmin expression = "Role-Admin"

	UiExprEnckey expression = "Crypto-PrivateKey"
	UiExprDeckey expression = "Crypto-PublicKey"
)

// db
const (
	ExprRegistercodeRecord expression = "Record-Registercode"
	ExprMarkdownItem       expression = "Record-Markdown-Item"
	ExprCalendarTable      expression = "Record-Calendar-Table"
	ExprNotificationItem   expression = "Record-Notification-Item"
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

// json field
// only expressions in the json body can be place here.
const (
	// user
	JsonExprAtomid       expression = "atomid"
	JsonExprUsername     expression = "username"
	JsonExprNickname     expression = "nickname"
	JsonExprPassword     expression = "password"
	JsonExprRegistercode expression = "registercode"
	JsonExprEntrycode    expression = "entrycode"
	JsonExprAccessToken  expression = "access_token"
	JsonExprRefreshToken expression = "refresh_token"
	JsonExprRole         expression = "role"

	// trace
	JsonExprTraceid   expression = "traceid"
	JsonExprCallChain expression = "call_chain"

	// calendar
	JsonExprCalendarID expression = "calendar_id"
	JsonExprRecords    expression = "records"
	JsonExprWeekday    expression = "weekday"
	JsonExprStartMin   expression = "start_min"
	JsonExprEndMin     expression = "end_min"
	JsonExprTitle      expression = "title"

	// notification
	JsonExprNid expression = "nid"
)

// json value
const (
	JsonExprMessageOk       expression = "ok"
	JsonExprMessageBadParam expression = "bad param"
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
	ModExprHandlerNotif    expression = "Handler-Notification"

	ModExprService      expression = "service"
	ModExprServiceUser  expression = "Service-User"
	ModExprServiceToken expression = "Service-Token"
	ModExprServiceAdmin expression = "Service-Admin"
	ModExprServiceNotif expression = "Service-Notification"

	ModExprMiddlware            expression = "middleware"
	ModExprMiddlwareAuthen      expression = "Middleware-Authen"
	ModExprMiddlwareInjector    expression = "Middleware-Injector"
	ModExprMiddlwareRateLimiter expression = "Middleware-RateLimiter"
	ModExprMiddlwareRoleControl expression = "Middleware-RoleControl"

	ModExprUtils expression = "Mod-Utils"
)

// type of request
const (
	// account
	ExprReqRegister  expression = "Request-Register"
	ExprReqLogin     expression = "Request-Login"
	ExprReqUserInfo  expression = "Request-UserInfo"
	ExprReqAdminSudo expression = "Request-AdminSudo"
	ExprReqRefresh   expression = "Request-Refresh"
	ExprReqSignout   expression = "Request-Signout"

	// calendar
	ExprReqCalendarGet    expression = "Request-CalendarGet"
	ExprReqCalendarPut    expression = "Request-CalendarPut"
	ExprReqCalendarDelete expression = "Request-CalendarDelete"

	// notif
	ExprReqNotifCreate         expression = "Request-NotificationCreate"
	ExprReqNotifUpdate         expression = "Request-NotificationUpdate"
	ExprReqNotifDelete         expression = "Request-NotificationDelete"
	ExprReqNotifWithdraw       expression = "Request-NotificationWithdraw"
	ExprReqNotifPost           expression = "Request-NotificationPost"
	ExprReqNotifGetSingle      expression = "Request-NotificationGetSingle"
	ExprReqNotifGetBatch       expression = "Request-NotificationGetBatch"
	ExprReqNotifGetUpdateBatch expression = "Request-NotificationGetUpdatebatch"
)

// router domain
const (
	DomainExprNull         expression = "Domain-Null"
	DomainExprPrivate      expression = "Domain-Private"
	DomainExprPublic       expression = "Domain-Public"
	DomainExprUnauthorized expression = "Domain-Unauthorized"
)
