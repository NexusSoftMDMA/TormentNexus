package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeout = 180 * time.Second
	httpTimeout    = 30 * time.Second
)

var http.DefaultClient = http.DefaultClient

// HandleExecuteCommand executes a raw command on the system
func HandleExecuteCommand(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	cmd := exec.Command("sh", "-c", command)
	cmd.Cancel = ctx.Done()

	done := make(chan struct{})
	var stdout, stderr strings.Builder

	go func() {
		out, e := cmd.CombinedOutput()
		if e == nil {
			stdout.Write(out)
		} else {
			stderr.Write(e.Error())
			if len(out) > 0 {
				stdout.Write(out)

		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		cmd.Process.Kill()
		return err("command timed out")
}

	result := map[string]interface{}{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"return_code": cmd.ProcessState.ExitCode(),
	}

	data, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(data))
}

}

// HandleNmapScan executes an Nmap scan against a target
func HandleNmapScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		return err("target is required")
}

	scanType, _ :=getString(args, "scan_type")
	if scanType == "" {
		scanType = "-sV"
	}
	ports, _ :=getString(args, "ports")
	additionalArgs, _ :=getString(args, "additional_args")

	cmdArgs := strings.Fields(scanType)
	if ports != "" {
		cmdArgs = append(cmdArgs, "-p", ports)

	cmdArgs = append(cmdArgs, strings.Fields(additionalArgs)...)
	cmdArgs = append(cmdArgs, target)

	cmd := exec.CommandContext(ctx, "nmap", cmdArgs...)
	cmd.Cancel = ctx.Done()

	done := make(chan struct{})
	var stdout, stderr strings.Builder

	go func() {
		out, e := cmd.CombinedOutput()
		if e == nil {
			stdout.Write(out)
		} else {
			stderr.Write(e.Error())
			if len(out) > 0 {
				stdout.Write(out)

		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		cmd.Process.Kill()
		return err("nmap scan timed out")
}

	result := map[string]interface{}{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"return_code": cmd.ProcessState.ExitCode(),
		"target":      target,
		"scan_type":   scanType,
	}

	data, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(data))
}

}
}

// HandleGobusterScan executes Gobuster to find directories, DNS subdomains, or virtual hosts
func HandleGobusterScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url is required")
}

	mode, _ :=getString(args, "mode")
	if mode == "" {
		mode = "dir"
	}
	wordlist, _ :=getString(args, "wordlist")
	if wordlist == "" {
		wordlist = "/usr/share/wordlists/dirb/common.txt"
	}
	additionalArgs, _ :=getString(args, "additional_args")

	cmdArgs := []string{mode, "-u", targetURL, "-w", wordlist}
	if additionalArgs != "" {
		cmdArgs = append(cmdArgs, strings.Fields(additionalArgs)...)

	cmd := exec.CommandContext(ctx, "gobuster", cmdArgs...)
	cmd.Cancel = ctx.Done()

	done := make(chan struct{})
	var stdout, stderr strings.Builder

	go func() {
		out, e := cmd.CombinedOutput()
		if e == nil {
			stdout.Write(out)
		} else {
			stderr.Write(e.Error())
			if len(out) > 0 {
				stdout.Write(out)

		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		cmd.Process.Kill()
		return err("gobuster scan timed out")
}

	result := map[string]interface{}{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"return_code": cmd.ProcessState.ExitCode(),
		"url":         targetURL,
		"mode":        mode,
	}

	data, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(data))
}

}
}

// HandleDirbScan executes Dirb web content scanner
func HandleDirbScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url is required")
}

	wordlist, _ :=getString(args, "wordlist")
	if wordlist == "" {
		wordlist = "/usr/share/wordlists/dirb/common.txt"
	}
	additionalArgs, _ :=getString(args, "additional_args")

	cmdArgs := []string{targetURL, wordlist}
	if additionalArgs != "" {
		cmdArgs = append(cmdArgs, strings.Fields(additionalArgs)...)

	cmd := exec.CommandContext(ctx, "dirb", cmdArgs...)
	cmd.Cancel = ctx.Done()

	done := make(chan struct{})
	var stdout, stderr strings.Builder

	go func() {
		out, e := cmd.CombinedOutput()
		if e == nil {
			stdout.Write(out)
		} else {
			stderr.Write(e.Error())
			if len(out) > 0 {
				stdout.Write(out)

		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		cmd.Process.Kill()
		return err("dirb scan timed out")
}

	result := map[string]interface{}{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"return_code": cmd.ProcessState.ExitCode(),
		"url":         targetURL,
	}

	data, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(data))
}

}
}

// HandleNiktoScan executes Nikto web server scanner
func HandleNiktoScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		return err("target is required")
}

	additionalArgs, _ :=getString(args, "additional_args")

	cmdArgs := []string{"-h", target}
	if additionalArgs != "" {
		cmdArgs = append(cmdArgs, strings.Fields(additionalArgs)...)

	cmd := exec.CommandContext(ctx, "nikto", cmdArgs...)
	cmd.Cancel = ctx.Done()

	done := make(chan struct{})
	var stdout, stderr strings.Builder

	go func() {
		out, e := cmd.CombinedOutput()
		if e == nil {
			stdout.Write(out)
		} else {
			stderr.Write(e.Error())
			if len(out) > 0 {
				stdout.Write(out)

		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		cmd.Process.Kill()
		return err("nikto scan timed out")
}

	result := map[string]interface{}{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"return_code": cmd.ProcessState.ExitCode(),
		"target":      target,
	}

	data, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(data))
}

}
}

// HandleHydraScan executes Hydra password cracking tool
func HandleHydraScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		return err("target is required")
}

	login, _ :=getString(args, "login")
	passList, _ :=getString(args, "password_list")
	if passList == "" {
		passList = "/usr/share/wordlists/rockyou.txt"
	}
	service, _ :=getString(args, "service")
	if service == "" {
		service = "ssh"
	}
	additionalArgs, _ :=getString(args, "additional_args")

	cmdArgs := []string{"-l", login, "-P", passList, "-t", "4", service + "://" + target}
	if additionalArgs != "" {
		cmdArgs = append(cmdArgs, strings.Fields(additionalArgs)...)

	cmd := exec.CommandContext(ctx, "hydra", cmdArgs...)
	cmd.Cancel = ctx.Done()

	done := make(chan struct{})
	var stdout, stderr strings.Builder

	go func() {
		out, e := cmd.CombinedOutput()
		if e == nil {
			stdout.Write(out)
		} else {
			stderr.Write(e.Error())
			if len(out) > 0 {
				stdout.Write(out)

		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		cmd.Process.Kill()
		return err("hydra scan timed out")
}

	result := map[string]interface{}{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"return_code": cmd.ProcessState.ExitCode(),
		"target":      target,
		"service":     service,
	}

	data, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(data))
}

}
}

// HandleSqlmapScan executes SQLMap SQL injection scanner
func HandleSqlmapScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url is required")
}

	additionalArgs, _ :=getString(args, "additional_args")

	cmdArgs := []string{"-u", targetURL, "--batch", "--random-agent"}
	if additionalArgs != "" {
		cmdArgs = append(cmdArgs, strings.Fields(additionalArgs)...)

	cmd := exec.CommandContext(ctx, "sqlmap", cmdArgs...)
	cmd.Cancel = ctx.Done()

	done := make(chan struct{})
	var stdout, stderr strings.Builder

	go func() {
		out, e := cmd.CombinedOutput()
		if e == nil {
			stdout.Write(out)
		} else {
			stderr.Write(e.Error())
			if len(out) > 0 {
				stdout.Write(out)

		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		cmd.Process.Kill()
		return err("sqlmap scan timed out")
}

	result := map[string]interface{}{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"return_code": cmd.ProcessState.ExitCode(),
		"url":         targetURL,
	}

	data, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(data))
}

}
}

// HandleWpscanScan executes WPScan WordPress vulnerability scanner
func HandleWpscanScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	targetURL, _ :=getString(args, "url")
	if targetURL == "" {
		return err("url is required")
}

	additionalArgs, _ :=getString(args, "additional_args")

	cmdArgs := []string{"--url", targetURL, "--enumerate", "vp,vt,tt,cb,dbe,u,m"}
	if additionalArgs != "" {
		cmdArgs = append(cmdArgs, strings.Fields(additionalArgs)...)

	cmd := exec.CommandContext(ctx, "wpscan", cmdArgs...)
	cmd.Cancel = ctx.Done()

	done := make(chan struct{})
	var stdout, stderr strings.Builder

	go func() {
		out, e := cmd.CombinedOutput()
		if e == nil {
			stdout.Write(out)
		} else {
			stderr.Write(e.Error())
			if len(out) > 0 {
				stdout.Write(out)

		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		cmd.Process.Kill()
		return err("wpscan timed out")
}

	result := map[string]interface{}{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"return_code": cmd.ProcessState.ExitCode(),
		"url":         targetURL,
	}

	data, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(data))
}

}
}

// HandleEnum4linuxScan executes enum4linux for SMB enumeration
func HandleEnum4linuxScan(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	target, _ :=getString(args, "target")
	if target == "" {
		return err("target is required")
}

	additionalArgs, _ :=getString(args, "additional_args")

	cmdArgs := []string{"-a", target}
	if additionalArgs != "" {
		cmdArgs = append(cmdArgs, strings.Fields(additionalArgs)...)

	cmd := exec.CommandContext(ctx, "enum4linux", cmdArgs...)
	cmd.Cancel = ctx.Done()

	done := make(chan struct{})
	var stdout, stderr strings.Builder

	go func() {
		out, e := cmd.CombinedOutput()
		if e == nil {
			stdout.Write(out)
		} else {
			stderr.Write(e.Error())
			if len(out) > 0 {
				stdout.Write(out)

		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		cmd.Process.Kill()
		return err("enum4linux scan timed out")
}

	result := map[string]interface{}{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"return_code": cmd.ProcessState.ExitCode(),
		"target":      target,
	}

	data, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(data))
}

}
}

// HandleJohnRipper executes John the Ripper password cracker
func HandleJohnRipper(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	hashFile, _ :=getString(args, "hash_file")
	if hashFile == "" {
		return err("hash_file is required")
}

	wordlist, _ :=getString(args, "wordlist")
	if wordlist == "" {
		wordlist = "/usr/share/wordlists/rockyou.txt"
	}
	additionalArgs, _ :=getString(args, "additional_args")

	cmdArgs := []string{"--wordlist=" + wordlist, hashFile}
	if additionalArgs != "" {
		cmdArgs = append(cmdArgs, strings.Fields(additionalArgs)...)

	cmd := exec.CommandContext(ctx, "john", cmdArgs...)
	cmd.Cancel = ctx.Done()

	done := make(chan struct{})
	var stdout, stderr strings.Builder

	go func() {
		out, e := cmd.CombinedOutput()
		if e == nil {
			stdout.Write(out)
		} else {
			stderr.Write(e.Error())
			if len(out) > 0 {
				stdout.Write(out)

		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		cmd.Process.Kill()
		return err("john the ripper timed out")
}

	result := map[string]interface{}{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"return_code": cmd.ProcessState.ExitCode(),
		"hash_file":   hashFile,
	}

	data, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(string(data))
}

}
}

// HandleMsfconsole executes Metasploit Framework commands
func HandleMsfconsole(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	command, _ :=getString(args, "command")
	if command == "" {
		return err("command is required")
}

	msfCmd := exec.CommandContext(ctx, "msfconsole", "-q", "-x", command)
	msfCmd.Cancel = ctx.Done()

	done := make(chan struct{})
	var stdout, stderr strings.Builder

	go func() {
		out, e := msfCmd.CombinedOutput()
		if e == nil {
			stdout.Write(out)
		} else {
			stderr.Write(e.Error())
			if len(out) > 0 {
				stdout.Write(out)

		}
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		msfCmd.Process.Kill()
		return err("msfconsole timed out")
}

	result := map[string]interface{}{
		"stdout":      stdout.String(),
		"stderr":      stderr.String(),
		"return_code": msfCmd.ProcessState.ExitCode(),
		"command":     command,
	}

	data, parseErr := json.Marshal(result)
	if parseErr != nil {
		return err(parseErr.Error())

}
}
}