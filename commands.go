package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

var gHeaders = make(map[string]string)

func getMethodKeys(node *TreeNode) []string {
	var keys []string
	for m := range node.Methods {
		keys = append(keys, strings.ToLower(m))
	}
	sort.Strings(keys)
	return keys
}

func lsImpl() {
	if len(gCurrentNode.Methods) > 0 {
		methods := getMethodKeys(gCurrentNode)
		fmt.Printf(".    [%s]\n", strings.Join(methods, "|"))
	} else {
		fmt.Println(".")
	}
	if gCurrentNode.Parent != nil {
		fmt.Println("..")
	}

	var childNames []string
	for name := range gCurrentNode.Children {
		childNames = append(childNames, name)
	}
	sort.Strings(childNames)

	for _, name := range childNames {
		child := gCurrentNode.Children[name]
		methods := getMethodKeys(child)
		if len(methods) > 0 {
			fmt.Printf("%s    [%s]\n", name, strings.Join(methods, "|"))
		} else {
			fmt.Printf("%s\n", name)
		}
	}
	fmt.Println("")
}

func cdImpl(args []string, root *TreeNode) {
	if len(args) == 1 { // no argument -> go to root
		gLabel = "/"
		gCurrentNode = root
		return
	}
	dest := args[1]
	if dest == "." {
		return
	}
	if dest == ".." {
		if gCurrentNode.Parent == nil {
			return
		}
		gLabel = gLabel[:strings.LastIndex(gLabel, "/")]
		gCurrentNode = gCurrentNode.Parent
		if gLabel == "" { // add / when cd to root
			gLabel = "/"
		}
		return
	}
	label := gLabel
	if label[len(label)-1] != '/' {
		label += "/"
	}
	label += dest
	for _, d := range gCurrentNode.Children {
		if d.Name == dest {
			gLabel = label
			gCurrentNode = d
			return
		}
	}
	msg := fmt.Sprintf("Warning: The '%v' endpoint is not present in the OpenAPI description", label)
	fmt.Println("\x1b[33m" + msg + ".\x1b[0m\n")
}

func highlightJSON(jsonStr string) string {
	var result strings.Builder
	inQuote := false
	var currentWord strings.Builder
	runes := []rune(jsonStr)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '"' {
			if inQuote {
				isKey := false
				for j := i + 1; j < len(runes); j++ {
					if runes[j] == ' ' || runes[j] == '\t' || runes[j] == '\n' || runes[j] == '\r' {
						continue
					}
					if runes[j] == ':' {
						isKey = true
					}
					break
				}
				currentWord.WriteRune(r)
				if isKey {
					result.WriteString("\x1b[33m" + currentWord.String() + "\x1b[0m") // Yellow for keys
				} else {
					result.WriteString("\x1b[32m" + currentWord.String() + "\x1b[0m") // Green for strings
				}
				currentWord.Reset()
				inQuote = false
			} else {
				inQuote = true
				currentWord.WriteRune(r)
			}
			continue
		}

		if inQuote {
			currentWord.WriteRune(r)
			continue
		}

		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
			currentWord.WriteRune(r)
			nextIsWordChar := false
			if i+1 < len(runes) {
				nr := runes[i+1]
				if (nr >= 'a' && nr <= 'z') || (nr >= '0' && nr <= '9') || nr == '-' || nr == '.' {
					nextIsWordChar = true
				}
			}
			if !nextIsWordChar {
				result.WriteString("\x1b[35m" + currentWord.String() + "\x1b[0m") // Magenta for other literals
				currentWord.Reset()
			}
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func prettyJSON(body []byte) {
	var parsedJSON interface{}
	err := json.Unmarshal(body, &parsedJSON)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	pretty, err := json.MarshalIndent(parsedJSON, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(highlightJSON(string(pretty)))
}

func printHeader(resp *http.Response) {
	statusColor := "\x1b[0m"
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		statusColor = "\x1b[32m" // Green
	} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		statusColor = "\x1b[33m" // Yellow
	} else if resp.StatusCode >= 400 {
		statusColor = "\x1b[31m" // Red
	}
	fmt.Printf("%s%s\x1b[0m\n", statusColor, resp.Status)

	var keys []string
	for k := range resp.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := resp.Header[k]
		fmt.Printf("\x1b[1;36m%v\x1b[0m: %v\n", k, v[0])
	}
}

func getBodyInteractive() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		tmpFile, err := os.CreateTemp("", "httprepl-body-*.json")
		if err != nil {
			return "", err
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpPath)

		cmd := exec.Command(editor, tmpPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return "", fmt.Errorf("failed to run editor: %w", err)
		}

		bodyBytes, err := os.ReadFile(tmpPath)
		if err != nil {
			return "", err
		}
		return string(bodyBytes), nil
	}

	fmt.Println("No EDITOR environment variable set. Enter request body (press Ctrl+D / Ctrl+Z or send EOF to finish):")
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return strings.Join(lines, "\n"), nil
}

func makeRequest(method string, args []string) {
	subPath := ""
	content := ""
	contentProvided := false

	for i := 1; i < len(args); i++ {
		if args[i] == "--content" || args[i] == "-c" {
			if i+1 < len(args) {
				content = args[i+1]
				contentProvided = true
				i++
			} else {
				fmt.Println("\x1b[31mError: Missing value for --content / -c flag\x1b[0m")
				return
			}
		} else {
			if strings.HasPrefix(args[i], "-") {
				fmt.Printf("\x1b[31mError: Unknown option %s\x1b[0m\n", args[i])
				return
			}
			if subPath != "" {
				fmt.Println("\x1b[31mError: Multiple sub-paths specified\x1b[0m")
				return
			}
			subPath = args[i]
		}
	}

	pathPart := subPath
	queryPart := ""
	if idx := strings.Index(subPath, "?"); idx != -1 {
		pathPart = subPath[:idx]
		queryPart = subPath[idx:]
	}

	u := baseAddr
	if !strings.HasSuffix(u, "/") && !strings.HasPrefix(gLabel, "/") {
		u += "/"
	}
	u += gLabel
	if pathPart != "" {
		if !strings.HasSuffix(u, "/") && !strings.HasPrefix(pathPart, "/") {
			u += "/"
		}
		u += pathPart
	}
	u += queryPart

	var reqBody io.Reader
	if method == "POST" || method == "PUT" || method == "PATCH" {
		if !contentProvided {
			var err error
			content, err = getBodyInteractive()
			if err != nil {
				fmt.Printf("\x1b[31mError getting request body: %v\x1b[0m\n", err)
				return
			}
		}
		reqBody = strings.NewReader(content)
	}

	req, err := http.NewRequest(method, u, reqBody)
	if err != nil {
		fmt.Printf("\x1b[31mError creating request: %v\x1b[0m\n", err)
		return
	}

	for k, v := range gHeaders {
		req.Header.Set(k, v)
	}
	if reqBody != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	fmt.Printf("Sending %s request to %s...\n", method, u)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("\x1b[31mError: Failed to get response: %v\x1b[0m\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	printHeader(resp)
	if len(body) > 0 {
		if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "json") {
			prettyJSON(body)
		} else {
			fmt.Println(string(body))
		}
	}
}

func handleHeaderCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: set header [name [value]]")
		return
	}
	if len(args) == 2 {
		if len(gHeaders) == 0 {
			fmt.Println("No headers set.")
		} else {
			for k, v := range gHeaders {
				fmt.Printf("%s: %s\n", k, v)
			}
		}
		return
	}
	key := args[2]
	if len(args) == 3 {
		delete(gHeaders, key)
		fmt.Printf("Header '%s' cleared.\n", key)
		return
	}
	value := strings.Join(args[3:], " ")
	gHeaders[key] = value
	fmt.Printf("Header '%s' set to '%s'.\n", key, value)
}

func handleClearCommand(args []string) {
	if len(args) < 3 || args[1] != "header" {
		fmt.Println("Usage: clear header <name>")
		return
	}
	key := args[2]
	delete(gHeaders, key)
	fmt.Printf("Header '%s' cleared.\n", key)
}

func getBaseURL() *url.URL {
	u, err := url.Parse(baseAddr)
	if err != nil {
		u, _ = url.Parse("http://localhost")
	}
	return u
}

func handleCookieCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: set cookie [name [value]]")
		return
	}
	if len(args) == 2 {
		showCookiesImpl()
		return
	}
	name := args[2]
	if len(args) == 3 {
		clearCookie(name)
		fmt.Printf("Cookie '%s' cleared.\n", name)
		return
	}
	value := strings.Join(args[3:], " ")

	jar := http.DefaultClient.Jar
	if jar == nil {
		fmt.Println("Error: Cookie jar is not initialized.")
		return
	}

	u := getBaseURL()
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	}

	jar.SetCookies(u, []*http.Cookie{cookie})
	fmt.Printf("Cookie '%s' set to '%s'.\n", name, value)
}

func clearCookie(name string) {
	jar := http.DefaultClient.Jar
	if jar == nil {
		return
	}
	u := getBaseURL()
	cookie := &http.Cookie{
		Name:   name,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	jar.SetCookies(u, []*http.Cookie{cookie})
}

func handleClearCookieCommand(args []string) {
	if len(args) < 3 || args[1] != "cookie" {
		fmt.Println("Usage: clear cookie <name>")
		return
	}
	name := args[2]
	clearCookie(name)
	fmt.Printf("Cookie '%s' cleared.\n", name)
}

func showCookiesImpl() {
	jar := http.DefaultClient.Jar
	if jar == nil {
		fmt.Println("Cookie jar is not initialized.")
		return
	}
	u := getBaseURL()
	cookies := jar.Cookies(u)
	if len(cookies) == 0 {
		fmt.Println("No cookies set.")
		return
	}
	for _, c := range cookies {
		fmt.Printf("%s: %s\n", c.Name, c.Value)
	}
}

func helpImpl() {
	fmt.Println("Available Commands:")
	fmt.Println("  ls                   List directory (endpoints and HTTP verbs)")
	fmt.Println("  cd [path]            Change directory (navigate to a path or sub-path; use '..' to go up)")
	fmt.Println("  tree                 Print the path tree structure")
	fmt.Println("  set header [k [v]]   Set or clear a custom header, or list all custom headers")
	fmt.Println("  clear header <k>     Clear a custom header")
	fmt.Println("  set cookie [k [v]]   Set or clear a cookie, or list all cookies")
	fmt.Println("  clear cookie <k>     Clear a cookie")
	fmt.Println("  show cookies         Show all cookies currently in the jar")
	fmt.Println("  get [sub-path]       Perform a GET request")
	fmt.Println("  post [sub-path]      Perform a POST request")
	fmt.Println("  put [sub-path]       Perform a PUT request")
	fmt.Println("  delete [sub-path]    Perform a DELETE request")
	fmt.Println("  patch [sub-path]     Perform a PATCH request")
	fmt.Println("  head [sub-path]      Perform a HEAD request")
	fmt.Println("  options [sub-path]   Perform an OPTIONS request")
	fmt.Println("  help                 Show this help menu")
	fmt.Println("  exit                 Exit the REPL")
}

func defaultCommand() {
	spew.Dump(gLabel)
	fmt.Println("\x1b[31mNo matching command found.\x1b[0m")
	fmt.Println("\x1b[31mExecute 'help' to see available commands.\x1b[0m")
}
