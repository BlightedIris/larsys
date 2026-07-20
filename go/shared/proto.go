package shared

type Action struct {
	NAME string
	PARAMS any
}

var REGISTER = Action{
	NAME: "register",
	PARAMS: struct{
		CLIENT_NAME string
	}{},
}

var REVOKE = Action{
	NAME: "revoke",
	PARAMS: struct{
		CLIENT_NAME string
	}{},
}

var PLUGIN_INSTALL = Action{
	NAME: "plugin/install",
	PARAMS: struct{
		NAME string
	}{},
}

var PLUGIN_REMOVE = Action{
	NAME: "plugin/remove",
	PARAMS: struct{
		NAME string
	}{},
}

type Request struct {
	SRC string
	TOKEN string
	ACTION Action
}

type Response struct {
	STATUS int
	MSG string
}

var OK = Response {
	STATUS : 0,
	MSG : "OK",
}

var UNAUTHORISED = Response {
	STATUS : 1,
	MSG : "Unauthorised",
}

var ALREADY_REGISTERED = Response {
	STATUS : 1,
	MSG : "Client already registered",
}
