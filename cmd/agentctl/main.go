package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/neurosai/agentos/pkg/version"
)

const defaultBase = "http://localhost:8080"
const defaultToken = "dev-token"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "version":
		fmt.Printf("%s %s\n", version.Name, version.Version)
	case "status":
		runStatus(os.Args[2:])
	case "task":
		runTask(os.Args[2:])
	case "audit":
		runAudit(os.Args[2:])
	case "tool":
		runTool(os.Args[2:])
	case "memory":
		runMemory(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func runStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	base := fs.String("base", defaultBase, "agentosd base URL")
	_ = fs.Parse(args)
	resp, err := http.Get(*base + "/readyz")
	if err != nil {
		fmt.Fprintf(os.Stderr, "status: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("readyz: %d %s\n", resp.StatusCode, strings.TrimSpace(string(body)))
}

func runTask(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: agentctl task <create|get|events|approve>")
		os.Exit(1)
	}
	switch args[0] {
	case "create":
		taskCreate(args[1:])
	case "get":
		taskGet(args[1:])
	case "events":
		taskEvents(args[1:])
	case "approve":
		taskApprove(args[1:])
	default:
		os.Exit(1)
	}
}

func taskCreate(args []string) {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	file := fs.String("f", "", "task yaml file")
	base := fs.String("base", defaultBase, "base URL")
	_ = fs.Parse(args)
	data, err := os.ReadFile(*file)
	if err != nil {
		fatal(err)
	}
	var doc struct {
		AgentRef  string            `yaml:"agentRef"`
		ContextID string            `yaml:"contextId"`
		Input     map[string]any    `yaml:"input"`
		Labels    map[string]string `yaml:"labels"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"agentRef":  doc.AgentRef,
		"contextId": doc.ContextID,
		"input":     doc.Input,
		"labels":    doc.Labels,
	})
	resp, err := doJSON("POST", *base+"/v1/tasks", body)
	if err != nil {
		fatal(err)
	}
	printJSON(resp)
}

func taskGet(args []string) {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	base := fs.String("base", defaultBase, "base URL")
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		os.Exit(1)
	}
	resp, err := doJSON("GET", *base+"/v1/tasks/"+fs.Arg(0), nil)
	if err != nil {
		fatal(err)
	}
	printJSON(resp)
}

func taskEvents(args []string) {
	fs := flag.NewFlagSet("events", flag.ExitOnError)
	base := fs.String("base", defaultBase, "base URL")
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		os.Exit(1)
	}
	req, _ := http.NewRequest("GET", *base+"/v1/tasks/"+fs.Arg(0)+"/events", nil)
	req.Header.Set("Authorization", "Bearer "+defaultToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fatal(err)
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		fmt.Println(sc.Text())
	}
}

func taskApprove(args []string) {
	fs := flag.NewFlagSet("approve", flag.ExitOnError)
	base := fs.String("base", defaultBase, "base URL")
	approved := fs.Bool("approved", true, "approval decision")
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		os.Exit(1)
	}
	body, _ := json.Marshal(map[string]any{"approved": *approved})
	resp, err := doJSON("POST", *base+"/v1/tasks/"+fs.Arg(0)+"/approvals", body)
	if err != nil {
		fatal(err)
	}
	printJSON(resp)
}

func runAudit(args []string) {
	if len(args) == 0 || args[0] != "trace" {
		os.Exit(1)
	}
	fs := flag.NewFlagSet("trace", flag.ExitOnError)
	base := fs.String("base", defaultBase, "base URL")
	_ = fs.Parse(args[1:])
	if fs.NArg() < 1 {
		os.Exit(1)
	}
	resp, err := doJSON("GET", *base+"/v1/audit/trace/"+fs.Arg(0), nil)
	if err != nil {
		fatal(err)
	}
	printJSON(resp)
}

func runTool(args []string) {
	if len(args) == 0 || args[0] != "invoke" {
		fmt.Fprintln(os.Stderr, "usage: agentctl tool invoke <toolId> [flags]")
		os.Exit(1)
	}
	fs := flag.NewFlagSet("invoke", flag.ExitOnError)
	base := fs.String("base", defaultBase, "base URL")
	taskID := fs.String("task", "", "task id")
	agentID := fs.String("agent", "agent:dev", "agent id")
	idem := fs.String("idempotency-key", "", "idempotency key")
	message := fs.String("arg", "", "shorthand message argument")
	title := fs.String("title", "", "shorthand title argument")
	_ = fs.Parse(args[1:])
	if fs.NArg() < 1 {
		os.Exit(1)
	}
	toolID := fs.Arg(0)
	argsMap := map[string]any{}
	if *message != "" {
		argsMap["message"] = *message
	}
	if *title != "" {
		argsMap["title"] = *title
	}
	body, _ := json.Marshal(map[string]any{
		"taskId":         *taskID,
		"agentId":        *agentID,
		"arguments":      argsMap,
		"idempotencyKey": *idem,
	})
	resp, err := doJSON("POST", *base+"/v1/tools/"+toolID+":invoke", body)
	if err != nil {
		fatal(err)
	}
	printJSON(resp)
}

func runMemory(args []string) {
	if len(args) == 0 {
		os.Exit(1)
	}
	switch args[0] {
	case "put":
		memoryPut(args[1:])
	case "search":
		memorySearch(args[1:])
	default:
		os.Exit(1)
	}
}

func memoryPut(args []string) {
	fs := flag.NewFlagSet("put", flag.ExitOnError)
	file := fs.String("f", "", "memory json file")
	base := fs.String("base", defaultBase, "base URL")
	_ = fs.Parse(args)
	if *file == "" && fs.NArg() > 0 {
		*file = fs.Arg(0)
	}
	data, err := os.ReadFile(*file)
	if err != nil {
		fatal(err)
	}
	resp, err := doJSON("POST", *base+"/v1/memory/records", data)
	if err != nil {
		fatal(err)
	}
	printJSON(resp)
}

func memorySearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	base := fs.String("base", defaultBase, "base URL")
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		os.Exit(1)
	}
	body, _ := json.Marshal(map[string]any{"query": fs.Arg(0)})
	resp, err := doJSON("POST", *base+"/v1/memory/query", body)
	if err != nil {
		fatal(err)
	}
	printJSON(resp)
}

func doJSON(method, url string, body []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+defaultToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return data, fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

func printJSON(data []byte) {
	var out bytes.Buffer
	_ = json.Indent(&out, data, "", "  ")
	fmt.Println(out.String())
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `usage:
  agentctl version
  agentctl status [--base URL]
  agentctl task create -f FILE
  agentctl task get <id>
  agentctl task events <id>
  agentctl task approve <id> [--approved]
  agentctl audit trace <trace-id>
  agentctl tool invoke <toolId> --task ID [--arg message=...]
  agentctl memory put -f FILE
  agentctl memory search <query>
`)
}
