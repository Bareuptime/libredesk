package main

import "github.com/zerodha/fastglue"

// handleAICompletion handles AI completion requests
func handleAICompletion(r *fastglue.Request) error {
	var (
		app       = r.Context.(*App)
		promptKey = string(r.RequestCtx.PostArgs().Peek("prompt_key"))
		content   = string(r.RequestCtx.PostArgs().Peek("content"))
	)
	resp, err := app.ai.Completion(promptKey, content)
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope(resp)
}

// handleGetAIPrompts returns AI prompts
func handleGetAIPrompts(r *fastglue.Request) error {
	var (
		app = r.Context.(*App)
	)
	resp, err := app.ai.GetPrompts()
	if err != nil {
		return sendErrorEnvelope(r, err)
	}
	return r.SendEnvelope(resp)
}
