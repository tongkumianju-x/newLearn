// mcp_demo.go
//
// 一个【自包含】的 MCP（Model Context Protocol）演示：
//   - 同一个二进制扮演两个角色：通过环境变量 MCP_ROLE=server 时是 MCP Server，
//     默认情况下是 MCP Client。
//   - Client 启动一份自己作为子进程（角色=server），通过 stdio 走 JSON-RPC 2.0 通信。
//   - Server 内置两个工具：add(a,b) 和 echo(text)，演示 tools/list 与 tools/call。
//
// 运行：
//   go run mcp_demo.go
//
// 协议要点：
//   - 传输：stdio（父进程读写子进程的 stdin/stdout）
//   - 报文：每行一个 JSON-RPC 2.0 消息（newline-delimited JSON）
//   - 流程：initialize -> notifications/initialized -> tools/list -> tools/call

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

// =====================================================================
// JSON-RPC 2.0 基础结构
// =====================================================================

type rpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"` // 通知没有 id
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// =====================================================================
// MCP 客户端
// =====================================================================

type MCPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader

	idGen   atomic.Int64
	mu      sync.Mutex
	pending map[int64]chan rpcMessage

	closed atomic.Bool
}

// NewMCPClient 启动一个 MCP Server 子进程
func NewMCPClient(command string, args ...string) (*MCPClient, error) {
	cmd := exec.Command(command, args...)
	// 让子进程以 server 角色运行
	cmd.Env = append(os.Environ(), "MCP_ROLE=server")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	// 关键修复：把 stderr 接到父进程，方便看到 server 的日志/错误
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	c := &MCPClient{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  bufio.NewReader(stdout),
		pending: make(map[int64]chan rpcMessage),
	}

	// 启动一个独立 goroutine 不停读消息，按 id 分发
	go c.readLoop()
	return c, nil
}

func (c *MCPClient) readLoop() {
	for {
		line, err := c.stdout.ReadBytes('\n')
		if err != nil {
			if !c.closed.Load() {
				fmt.Fprintln(os.Stderr, "[client] readLoop 退出:", err)
			}
			return
		}
		var msg rpcMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			fmt.Fprintln(os.Stderr, "[client] 收到非法 JSON:", string(line))
			continue
		}
		if msg.ID == nil {
			// 通知：本 demo 不处理
			continue
		}
		c.mu.Lock()
		ch, ok := c.pending[*msg.ID]
		if ok {
			delete(c.pending, *msg.ID)
		}
		c.mu.Unlock()
		if ok {
			ch <- msg
		}
	}
}

// send 发送请求并等待响应（带超时）
func (c *MCPClient) send(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	id := c.idGen.Add(1)
	ch := make(chan rpcMessage, 1)
	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	paramsRaw, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req := rpcMessage{JSONRPC: "2.0", ID: &id, Method: method, Params: paramsRaw}
	if err := c.writeMsg(req); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	}
}

// notify 发送一个通知（无需等待响应）
func (c *MCPClient) notify(method string, params interface{}) error {
	paramsRaw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return c.writeMsg(rpcMessage{JSONRPC: "2.0", Method: method, Params: paramsRaw})
}

func (c *MCPClient) writeMsg(v rpcMessage) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = c.stdin.Write(b)
	return err
}

func (c *MCPClient) Close() {
	c.closed.Store(true)
	_ = c.stdin.Close()
	_ = c.cmd.Process.Kill()
	_ = c.cmd.Wait()
}

// ---------- 高层 API ----------

func (c *MCPClient) Initialize(ctx context.Context) (json.RawMessage, error) {
	res, err := c.send(ctx, "initialize", map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo":      map[string]string{"name": "go-mcp-demo", "version": "0.0.1"},
	})
	if err != nil {
		return nil, err
	}
	if err := c.notify("notifications/initialized", map[string]interface{}{}); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *MCPClient) ListTools(ctx context.Context) (json.RawMessage, error) {
	return c.send(ctx, "tools/list", map[string]interface{}{})
}

func (c *MCPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (json.RawMessage, error) {
	return c.send(ctx, "tools/call", map[string]interface{}{
		"name":      name,
		"arguments": args,
	})
}

// =====================================================================
// MCP 服务端（内置，被同一个二进制以 MCP_ROLE=server 启动）
// =====================================================================

type toolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

func runServer() {
	in := bufio.NewReader(os.Stdin)
	out := os.Stdout

	tools := []toolDef{
		{
			Name:        "add",
			Description: "返回两个整数 a + b 的和",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{"type": "number"},
					"b": map[string]interface{}{"type": "number"},
				},
				"required": []string{"a", "b"},
			},
		},
		{
			Name:        "echo",
			Description: "原样返回传入的 text",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{"type": "string"},
				},
				"required": []string{"text"},
			},
		},
	}

	writeMsg := func(msg rpcMessage) {
		b, _ := json.Marshal(msg)
		b = append(b, '\n')
		_, _ = out.Write(b)
	}
	rawResult := func(v interface{}) json.RawMessage {
		b, _ := json.Marshal(v)
		return b
	}
	respondErr := func(id *int64, code int, message string) {
		writeMsg(rpcMessage{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: message}})
	}

	for {
		line, err := in.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Fprintln(os.Stderr, "[server] 读取失败:", err)
			}
			return
		}
		var msg rpcMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			fmt.Fprintln(os.Stderr, "[server] 非法 JSON:", string(line))
			continue
		}

		switch msg.Method {
		case "initialize":
			writeMsg(rpcMessage{
				JSONRPC: "2.0", ID: msg.ID,
				Result: rawResult(map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
					"serverInfo":      map[string]string{"name": "go-mcp-demo-server", "version": "0.0.1"},
				}),
			})

		case "notifications/initialized":
			// 通知，不回包
			fmt.Fprintln(os.Stderr, "[server] client 已完成 initialize")

		case "tools/list":
			writeMsg(rpcMessage{
				JSONRPC: "2.0", ID: msg.ID,
				Result: rawResult(map[string]interface{}{"tools": tools}),
			})

		case "tools/call":
			var p struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			if err := json.Unmarshal(msg.Params, &p); err != nil {
				respondErr(msg.ID, -32602, "invalid params")
				continue
			}
			text, ierr := invokeTool(p.Name, p.Arguments)
			if ierr != nil {
				respondErr(msg.ID, -32000, ierr.Error())
				continue
			}
			writeMsg(rpcMessage{
				JSONRPC: "2.0", ID: msg.ID,
				Result: rawResult(map[string]interface{}{
					"content": []map[string]interface{}{
						{"type": "text", "text": text},
					},
				}),
			})

		default:
			if msg.ID != nil {
				respondErr(msg.ID, -32601, "method not found: "+msg.Method)
			}
		}
	}
}

func invokeTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	case "add":
		a, ok1 := args["a"].(float64)
		b, ok2 := args["b"].(float64)
		if !ok1 || !ok2 {
			return "", fmt.Errorf("add 需要两个数字参数 a 和 b")
		}
		return fmt.Sprintf("%v", a+b), nil
	case "echo":
		text, ok := args["text"].(string)
		if !ok {
			return "", fmt.Errorf("echo 需要 string 类型参数 text")
		}
		return text, nil
	default:
		return "", fmt.Errorf("未知工具: %s", name)
	}
}

// =====================================================================
// main：根据 MCP_ROLE 切换角色
// =====================================================================

func main() {
	if os.Getenv("MCP_ROLE") == "server" {
		runServer()
		return
	}

	// 用 os.Args[0]（自身）启动一个 server 子进程
	self, err := os.Executable()
	if err != nil {
		fmt.Println("获取自身可执行路径失败:", err)
		return
	}

	client, err := NewMCPClient(self)
	if err != nil {
		fmt.Println("启动 MCP Server 失败:", err)
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. 握手
	initRes, err := client.Initialize(ctx)
	if err != nil {
		fmt.Println("❌ initialize 失败:", err)
		return
	}
	fmt.Println("✅ MCP Server 初始化成功")
	fmt.Println("   serverInfo:", string(initRes))

	// 2. 拉工具列表
	tools, err := client.ListTools(ctx)
	if err != nil {
		fmt.Println("❌ tools/list 失败:", err)
		return
	}
	fmt.Println("📦 可用工具列表:")
	prettyPrint(tools)

	// 3. 调用 add 工具
	res1, err := client.CallTool(ctx, "add", map[string]interface{}{"a": 12, "b": 30})
	if err != nil {
		fmt.Println("❌ tools/call add 失败:", err)
		return
	}
	fmt.Println("🧮 add(12, 30) =>")
	prettyPrint(res1)

	// 4. 调用 echo 工具
	res2, err := client.CallTool(ctx, "echo", map[string]interface{}{"text": "hello MCP!"})
	if err != nil {
		fmt.Println("❌ tools/call echo 失败:", err)
		return
	}
	fmt.Println("🔁 echo(\"hello MCP!\") =>")
	prettyPrint(res2)
}

func prettyPrint(raw json.RawMessage) {
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		fmt.Println(string(raw))
		return
	}
	b, _ := json.MarshalIndent(v, "   ", "  ")
	fmt.Println("  ", string(b))
}
