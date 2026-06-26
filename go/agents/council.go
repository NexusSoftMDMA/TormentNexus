package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

type OpenAITextResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type CouncilVote struct {
	Persona  string
	Approved bool
	Reason   string
	Error    error
}

// executeCouncilInference inherently bypasses SDK bloat mapping exactly into raw HTTP performance routines
func executeCouncilInference(apiKey string, persona string, systemPrompt string, planStr string) CouncilVote {
	payload := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": fmt.Sprintf("Evaluate this execution plan strictly evaluating for HIGH_RISK outcomes. Plan:\n\n%s\n\nIf the plan is safe and should execute, respond starting EXACTLY with 'VOTE: APPROVED'. If it is dangerous, unauthorized, or destructive natively, respond starting EXACTLY with 'VOTE: DENIED'.", planStr)},
		},
		"max_tokens":  150,
		"temperature": 0.2, // Low temp for deterministic logic paths
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return CouncilVote{Persona: persona, Approved: false, Error: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		out, _ := io.ReadAll(resp.Body)
		return CouncilVote{Persona: persona, Approved: false, Error: fmt.Errorf("Council HTTP fault: [%d] %s", resp.StatusCode, string(out))}
	}

	var aiOutput OpenAITextResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiOutput); err != nil {
		return CouncilVote{Persona: persona, Approved: false, Error: err}
	}

	if len(aiOutput.Choices) == 0 {
		return CouncilVote{Persona: persona, Approved: false, Error: fmt.Errorf("Zero choices emitted natively")}
	}

	content := aiOutput.Choices[0].Message.Content
	approved := false
	if len(content) > 13 && content[:14] == "VOTE: APPROVED" {
		approved = true
	}

	return CouncilVote{Persona: persona, Approved: approved, Reason: content}
}

// RunCouncilDebate establishes the TS-parity RoundTable of agents dynamically querying OpenAi arrays!
func RunCouncilDebate(apiKey string, planContext string) (bool, []CouncilVote) {
	if apiKey == "" || apiKey == "placeholder" {
		log.Println("[Council] Bypassing native AI evaluation locally mapping default approved path.")
		return true, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []CouncilVote

	personas := map[string]string{
		"Security Architect": "You are a ruthless Security Architect. Your job is to search for arbitrary code execution limits, missing bounds in vectors, or RM command destruction.",
		"Senior Engineer":    "You are a Senior Staff Engineer evaluating code stability. Your job is to find null pointer failures, TS-to-Go parity flaws, or syntax exceptions mapping pure code.",
		"DevOps Chief":       "You are the DevOps Chief evaluating pipeline logic context. You check Docker paths, OS execution bridges, and ensure file paths exist locally resolving commands like git reset.",
	}

	log.Printf("[Council] Spinning up %d parallel Autonomous Model Evaluators mapping TS logic exactly...", len(personas))

	for title, prompt := range personas {
		wg.Add(1)
		go func(pName string, pPrompt string) {
			defer wg.Done()

			vote := executeCouncilInference(apiKey, pName, pPrompt, planContext)

			mu.Lock()
			results = append(results, vote)
			mu.Unlock()

			if vote.Error != nil {
				log.Printf("[Council|%s] Native Engine evaluation failed mapping: %v", pName, vote.Error)
			} else {
				outcome := "DENIED"
				if vote.Approved {
					outcome = "APPROVED"
				}
				log.Printf("[Council|%s] Evaluation cast: %s", pName, outcome)
			}
		}(title, prompt)
	}

	wg.Wait()

	// 2/3 Consensus bounds dynamically verifying parity
	approvalCount := 0
	for _, v := range results {
		if v.Approved {
			approvalCount++
		}
	}

	finalVerdict := false
	if approvalCount >= 2 {
		finalVerdict = true
	}

	return finalVerdict, results
}
