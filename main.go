package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

type Check struct {
	Title   string
	Command string
	Timeout int // [sec]
	Outputs *[]Output
}

type Output struct {
	Filepath string
	Diff     string
}

func main() {
	check := Check{
		Title:   "go fmt",
		Command: "cd sample && go fmt .",
		Timeout: 10,
	}

	PanicIfError(Execute(check.Command, time.Duration(check.Timeout)*time.Second))
	PanicIfError(exec.Command("git", "add", "-A").Run())
	PanicIfError(exec.Command("git", "commit", "-m", check.Title).Run())

	outputs := []Output{}
	out := execOut("git", "diff", "HEAD~..HEAD", "--name-only")
	files := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	for _, file := range files {
		out := execOut("git", "diff", "HEAD~..HEAD", file)
		outputs = append(outputs, Output{
			Filepath: file,
			Diff:     out,
		})
	}
	check.Outputs = &outputs

	fmt.Println(format(&check))
}

func format(check *Check) string {
	if len(*check.Outputs) == 0 {
		return ""
	}

	tmpl := `## diff for {{ .Title }}
{{ range .Outputs }}
<details>
<summary>{{ .Filepath }}</summary>` +
		"\n\n```diff\n{{ .Diff }}\n```" +
		`
</details>
{{ end }}`

	t, err := template.New("").Parse(tmpl)
	PanicIfError(err)
	var out bytes.Buffer
	PanicIfError(t.Execute(&out, check))

	return out.String()
}

// TODO: ちゃんとやる？
func execOut(name string, arg ...string) string {
	c := exec.Command(name, arg...)
	var stdout bytes.Buffer
	c.Stdout = &stdout
	PanicIfError(c.Run())
	return stdout.String()
}

func Execute(command string, timeout time.Duration) error {
	shell, ok := os.LookupEnv("SHELL")
	if !ok {
		shell = "sh"
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c := exec.CommandContext(ctx, shell, "-c", command)
	// TODO: stdout/stderrをリダイレクト
	return c.Run()
}

// TODO: ちゃんとやる
func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
