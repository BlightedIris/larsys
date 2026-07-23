package proto

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
	NAME    string
	VERSION string
}

var PLUGIN_INSTALL = Action{
	NAME:   "plugin/install",
	PARAMS: PLUGIN_INSTALL_PARAMS{},
}

type PLUGIN_DOWNLOAD_PARAMS struct {
	NAME string
	FILE []byte
}

type PLUGIN_UPLOAD_PARAMS struct {
	NAME string
	FILE []byte
}

var PLUGIN_UPLOAD = Action{
	NAME:   "plugin/upload",
	PARAMS: PLUGIN_UPLOAD_PARAMS{},
}

type PLUGIN_UNINSTALL_PARAMS struct {
	NAME string
}

var PLUGIN_UNINSTALL = Action{
	NAME:   "plugin/uninstall",
	PARAMS: PLUGIN_UNINSTALL_PARAMS{},
}
