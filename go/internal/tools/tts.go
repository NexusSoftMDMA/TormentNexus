//go:build ignore
// +build ignore

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// HandleSayTTS speaks text aloud using the host OS native text-to-speech engine.
func HandleSayTTS(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ := getString(args, "text")
	if text == "" {
		return err("text parameter is required")
	}

	rate := getInt(args, "rate")
	voice, _ := getString(args, "voice")

	switch runtime.GOOS {
	case "darwin":
		// macOS native say command
		cmdArgs := []string{text}
		if voice != "" {
			cmdArgs = append(cmdArgs, "-v", voice)
		}
		if rate > 0 {
			cmdArgs = append(cmdArgs, "-r", fmt.Sprintf("%d", rate))
		}
		cmd := exec.CommandContext(ctx, "say", cmdArgs...)
		if out, e := cmd.CombinedOutput(); e != nil {
			return err(fmt.Sprintf("macOS say failed: %v, output: %s", e, string(out)))
		}
		return ok("Successfully synthesized speech on macOS.")

	case "windows":
		// Windows PowerShell SpeechSynthesis
		// Basic command: Add-Type -AssemblyName System.Speech; (New-Object System.Speech.Synthesis.SpeechSynthesizer).Speak("text")
		psCmd := fmt.Sprintf(`Add-Type -AssemblyName System.Speech; $synth = New-Object System.Speech.Synthesis.SpeechSynthesizer;`)
		if voice != "" {
			psCmd += fmt.Sprintf(` $synth.SelectVoice(%q);`, voice)
		}
		if rate > 0 {
			// SAPI rate ranges from -10 to 10
			// Map rate words-per-minute (around 200) to SAPI scale
			sapiRate := (rate - 200) / 10
			if sapiRate < -10 {
				sapiRate = -10
			}
			if sapiRate > 10 {
				sapiRate = 10
			}
			psCmd += fmt.Sprintf(` $synth.Rate = %d;`, sapiRate)
		}
		psCmd += fmt.Sprintf(` $synth.Speak(%q)`, text)

		cmd := exec.CommandContext(ctx, "powershell", "-Command", psCmd)
		if out, e := cmd.CombinedOutput(); e != nil {
			return err(fmt.Sprintf("Windows speech synthesis failed: %v, output: %s", e, string(out)))
		}
		return ok("Successfully synthesized speech on Windows.")

	case "linux":
		// Linux espeak fallback
		cmdArgs := []string{text}
		if voice != "" {
			cmdArgs = append(cmdArgs, "-v", voice)
		}
		if rate > 0 {
			cmdArgs = append(cmdArgs, "-s", fmt.Sprintf("%d", rate))
		}
		cmd := exec.CommandContext(ctx, "espeak", cmdArgs...)
		if out, e := cmd.CombinedOutput(); e != nil {
			// Try spd-say
			cmdSpd := exec.CommandContext(ctx, "spd-say", text)
			if e2 := cmdSpd.Run(); e2 != nil {
				return err(fmt.Sprintf("Linux speech synthesis failed. Tested both espeak and spd-say. error: %v, espeak output: %s", e, string(out)))
			}
		}
		return ok("Successfully synthesized speech on Linux.")

	default:
		return err(fmt.Sprintf("Text-to-speech is not supported on host OS: %s", runtime.GOOS))
	}
}

// HandleOpenAITTS converts text to speech using OpenAI API and saves the file.
func HandleOpenAITTS(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return err("OPENAI_API_KEY environment variable is not set")
	}

	text, _ := getString(args, "text")
	if text == "" {
		return err("text parameter is required")
	}

	voice, _ := getString(args, "voice")
	if voice == "" {
		voice = "alloy"
	}

	model, _ := getString(args, "model")
	if model == "" {
		model = "tts-1"
	}

	speed := 1.0
	if speedVal, exists := args["speed"]; exists {
		if f, okF := speedVal.(float64); okF {
			speed = f
		}
	}

	// Build OpenAI request
	reqBody := map[string]interface{}{
		"model":          model,
		"input":          text,
		"voice":          voice,
		"speed":          speed,
		"response_format": "mp3",
	}

	jsonData, errMarshal := json.Marshal(reqBody)
	if errMarshal != nil {
		return errResponseTTS(errMarshal)
	}

	apiURL := os.Getenv("OPENAI_API_URL")
	if apiURL == "" {
		apiURL = "https://api..com/v1/audio/speech"
	}
	req, errNew := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if errNew != nil {
		return errResponseTTS(errNew)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, errDo := client.Do(req)
	if errDo != nil {
		return errResponseTTS(errDo)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("OpenAI API returned status %d: %s", resp.StatusCode, string(respBytes)))
	}

	// Save to temp file
	tempFile, errTemp := os.CreateTemp("", "-tts-*.mp3")
	if errTemp != nil {
		return errResponseTTS(errTemp)
	}
	defer tempFile.Close()

	if _, errCopy := io.Copy(tempFile, resp.Body); errCopy != nil {
		return errResponseTTS(errCopy)
	}

	absPath, _ := filepath.Abs(tempFile.Name())
	return ok(fmt.Sprintf("Speech successfully synthesized and saved to: %s", absPath))
}

func errResponseTTS(err error) (ToolResponse, error) {
	return ToolResponse{
		Content: []TextContent{{Type: "text", Text: fmt.Sprintf("TTS failed: %v", err)}},
		IsError: true,
	}, nil
}
