package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/goclone-dev/goclone/pkg/crawler"
	"github.com/goclone-dev/goclone/pkg/file"
	"github.com/goclone-dev/goclone/pkg/html"
	"github.com/goclone-dev/goclone/pkg/parser"
	"github.com/goclone-dev/goclone/pkg/server"
)

// CloneOptions contains all the options for the cloning process
type CloneOptions struct {
	Serve     bool
	Open      bool
	ServePort int
	Cookies   []string
	Proxy     string
	UserAgent string
}

// CloneSite clones the site with the specified options
func CloneSite(ctx context.Context, args []string, opts CloneOptions) error {
	jar, err := setupCookieJar(args, opts.Cookies)
	if err != nil {
		return err
	}

	firstProject, err := cloneProjects(ctx, args, jar, opts)
	if err != nil {
		return err
	}

	return handlePostCloneActions(ctx, firstProject, opts)
}

func setupCookieJar(args []string, cookies []string) (*cookiejar.Jar, error) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}

	if len(cookies) == 0 {
		return jar, nil
	}

	cs := make([]*http.Cookie, 0, len(cookies))
	for _, c := range cookies {
		ff := strings.Fields(c)
		for _, f := range ff {
			var k, v string
			if i := strings.IndexByte(f, '='); i >= 0 {
				k, v = f[:i], strings.TrimRight(f[i+1:], ";")
			} else {
				return nil, fmt.Errorf("no = in cookie %q", c)
			}
			cs = append(cs, &http.Cookie{Name: k, Value: v})
		}
	}

	for _, a := range args {
		u, err := url.Parse(a)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", a, err)
		}
		jar.SetCookies(&url.URL{Scheme: u.Scheme, User: u.User, Host: u.Host}, cs)
	}

	return jar, nil
}

func cloneProjects(ctx context.Context, args []string, jar *cookiejar.Jar, opts CloneOptions) (string, error) {
	var firstProject string
	for _, u := range args {
		isValid, isValidDomain := parser.ValidateURL(u), parser.ValidateDomain(u)
		if !isValid && !isValidDomain {
			return "", fmt.Errorf("%q is not valid", u)
		}

		name := u
		if isValidDomain {
			u = parser.CreateURL(name)
		} else {
			name = parser.GetDomain(u)
		}

		projectPath := file.CreateProject(name)
		if firstProject == "" {
			firstProject = projectPath
		}

		if err := crawler.Crawl(ctx, u, projectPath, jar, opts.Proxy, opts.UserAgent); err != nil {
			return "", fmt.Errorf("%q: %w", u, err)
		}

		if err := html.LinkRestructure(projectPath); err != nil {
			return "", fmt.Errorf("%q: %w", projectPath, err)
		}
	}
	return firstProject, nil
}

func handlePostCloneActions(ctx context.Context, projectPath string, opts CloneOptions) error {
	if opts.Serve {
		// Start the server in a goroutine
		go func() {
			if err := server.Serve(projectPath, opts.ServePort); err != nil {
				fmt.Printf("Error starting server: %v\n", err)
			}
		}()

		// Wait a moment to ensure the server is ready
		time.Sleep(100 * time.Millisecond)

		if opts.Open {
			if err := openInBrowser(opts.ServePort); err != nil {
				return err
			}
		}

		// Wait for context cancellation (Ctrl+C)
		<-ctx.Done()
		return nil
	}

	if opts.Open {
		return openFile(projectPath + "/index.html")
	}

	return nil
}

func openInBrowser(port int) error {
	cmd := open("http://localhost:" + fmt.Sprintf("%d", port))
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%v: %w", cmd.Args, err)
	}
	return nil
}

func openFile(path string) error {
	cmd := open(path)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%v: %w", cmd.Args, err)
	}
	return nil
}

// open opens the specified URL in the default browser of the user.
func open(url string) *exec.Cmd {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...)
}
