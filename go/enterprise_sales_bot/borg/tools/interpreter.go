package tools

import (
	"github.com/NexusSoftMDMA/TormentNexus/repl"
)

var sessions = make(map[string]*repl.Session)

func (r *Registry) registerInterpreterTools() {
	r.Tools = append(r.Tools, Tool{
		Name:        "code_interpreter",
		Description: "Executes code statefully in a persistent session. Arguments: language (string: 'python' or 'node'), code (string)",
		Execute: func(args map[string]interface{}) (string, error) {
			lang, _ := args["language"].(string)
			code, _ := args["code"].(string)

			session, ok := sessions[lang]
			if !ok {
				var err error
				session, err = repl.NewSession(lang)
				if err != nil {
					return "", err
				}
				sessions[lang] = session
			}

			return session.Execute(code)
		},
	})
}
