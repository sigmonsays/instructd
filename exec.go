package main

type ExecRequest struct {
	Values *MapVals

	Id     string            `json:"id"`
	Body   string            `json:"body"`
	Pwd    string            `json:"pwd"`
	StdOut bool              `json:"stdout"`
	StdErr bool              `json:"stderr"`
	Env    map[string]string `json:"env"`
	Args   []string          `json:"args"`
}

type ExecResponse struct {
	Error    string `json:"error"`
	ExitCode int    `json:"exit_code"`
	Pid      int    `json:"pid"`
	StdOut   string `json:"stdout"`
	StdErr   string `json:"stderr"`
	DurMs    int64  `json:"dur_ms"`
}
