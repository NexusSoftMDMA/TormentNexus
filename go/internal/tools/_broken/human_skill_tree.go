package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"
)

func HandleGetSkill(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    skillPath, _ :=getString(args, "skill_path")
    if skillPath == "" {
        return err("missing required parameter: skill_path")
}

    url := "https://raw.githubusercontent.com/24kchengYe/human-skill-tree/main/skills/" + skillPath + "/SKILL.md"
    client := http.Client{Timeout: 30 * time.Second}
    req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if reqErr != nil {
        return err("failed to create request: " + reqErr.Error())
}

    resp, fetchErr := client.Do(req)
    if fetchErr != nil {
        return err("failed to fetch skill: " + fetchErr.Error())
}

    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return err("skill not found or error: HTTP " + resp.Status)
}

    buf := new(strings.Builder)
    _, readErr := io.Copy(buf, resp.Body)
    if readErr != nil {
        return err("failed to read response: " + readErr.Error())
}

    return ok(buf.String())
}

func HandleListSkills(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url := "https://api.github.com/repos/24kchengYe/human-skill-tree/contents/skills"
    client := http.Client{Timeout: 30 * time.Second}
    req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if reqErr != nil {
        return err("failed to create request: " + reqErr.Error())
}

    resp, fetchErr := client.Do(req)
    if fetchErr != nil {
        return err("failed to list skills: " + fetchErr.Error())
}

    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return err("failed to list skills: HTTP " + resp.Status)
}

    var items []struct {
        Name string `json:"name"`
        Type string `json:"type"`
    }
    parseErr := json.NewDecoder(resp.Body).Decode(&items)
    if parseErr != nil {
        return err("failed to parse response: " + parseErr.Error())
}

    var result strings.Builder
    result.WriteString("Available skills (phase prefix - name):\n")
    for _, item := range items {
        if item.Type != "dir" {
            continue
        }
        phase := ""
        if len(item.Name) >= 2 {
            phaseNum := item.Name[:2]
            phaseNames := map[string]string{
                "00": "Learning How to Learn",
                "01": "K-12 Foundation",
                "02": "University",
                "03": "Graduate & Research",
                "04": "Career",
                "05": "Social Intelligence",
                "06": "Self-Development",
            }
            if p, found := phaseNames[phaseNum]; found {
                phase = p
            } else {
                phase = "Unknown Phase"
            }
        }
        skillName := item.Name
        if len(item.Name) > 3 {
            skillName = item.Name[3:]
        }
        result.WriteString(fmt.Sprintf("- %s (Phase: %s)\n", skillName, phase))

    return ok(result.String())
}
}