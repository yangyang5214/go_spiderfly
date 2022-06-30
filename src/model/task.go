package model

type /**/Task struct {
	EntryPoint       string
	Domain string
	EntryPointHost   string
	TargetDir        string
	Cookie        string
	ExtraHeaders     map[string]interface{}
}
