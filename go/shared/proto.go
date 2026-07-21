package shared

type Action struct {
	NAME   string
	PARAMS any
}

var PING = Action{
	NAME:   "ping",
	PARAMS: struct{}{},
}

type REGISTER_PARAMS struct {
	NAME string
}

var REGISTER = Action{
	NAME:   "register",
	PARAMS: REGISTER_PARAMS{},
}

type REVOKE_PARAMS struct {
	NAME string
}

var REVOKE = Action{
	NAME:   "revoke",
	PARAMS: REVOKE_PARAMS{},
}

type PLUGIN_INSTALL_PARAMS struct {
	NAME string
}

var PLUGIN_INSTALL = Action{
	NAME:   "plugin/install",
	PARAMS: PLUGIN_INSTALL_PARAMS{},
}

type PLUGIN_UNINSTALL_PARAMS struct {
	NAME string
}

var PLUGIN_UNINSTALL = Action{
	NAME:   "plugin/uninstall",
	PARAMS: PLUGIN_UNINSTALL_PARAMS{},
}

type Request struct {
	SRC    string
	TOKEN  string
	ACTION Action
}

type Response struct {
	SRC    string
	STATUS int
	MSG    string
	PARAMS any
}

type REGISTERED_PARAMS struct {
	TOKEN string
}

var REGISTERED = Response{
	STATUS: 0,
	MSG:    "+++ Client Registered +++",
	PARAMS: REGISTERED_PARAMS{},
}

var OK = Response{
	STATUS: 0,
	MSG:    "OK",
}

var UNAUTHORISED = Response{
	STATUS: 1,
	MSG:    "Unauthorised",
}

var ALREADY_REGISTERED = Response{
	STATUS: 1,
	MSG:    "Client already registered",
}
