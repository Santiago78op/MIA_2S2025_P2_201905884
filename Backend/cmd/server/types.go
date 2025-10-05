package main

// RunCommandRequest representa una solicitud para ejecutar un comando
type RunCommandRequest struct {
	Line string `json:"line"` // Línea de comando a ejecutar
}

// RunCommandResponse representa la respuesta de ejecutar un comando
type RunCommandResponse struct {
	OK      bool              `json:"ok"`
	Output  string            `json:"output,omitempty"`
	Error   string            `json:"error,omitempty"`
	Input   string            `json:"input,omitempty"`
	Command string            `json:"command,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
	Usage   string            `json:"usage,omitempty"`
}

// ScriptRequest representa una solicitud para ejecutar un script
type ScriptRequest struct {
	Script string `json:"script"` // Script con múltiples líneas de comandos
}

// ScriptResponse representa la respuesta de ejecutar un script
type ScriptResponse struct {
	OK           bool            `json:"ok"`
	Results      []CommandResult `json:"results"`
	Error        string          `json:"error,omitempty"`
	TotalLines   int             `json:"total_lines"`
	Executed     int             `json:"executed"`
	SuccessCount int             `json:"success_count"`
	ErrorCount   int             `json:"error_count"`
}

// CommandResult representa el resultado de ejecutar un comando individual
type CommandResult struct {
	Line    int    `json:"line"`
	Input   string `json:"input"`
	Output  string `json:"output,omitempty"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
