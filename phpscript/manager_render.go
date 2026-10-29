package phpscript

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/titpetric/phpscript/runner"
)

// ErrNotFound is returned when the path names no module, or a path under a
// module that its public tree does not hold and that index.php answered 404
// for.
var ErrNotFound = errors.New("no module there")

// Render answers one request against a module: a file under its public
// directory as bytes, and anything else as whatever index.php echoed.
//
// The request is not rewritten, so $_SERVER["REQUEST_URI"] holds the path the
// browser requested, prefix and all.
func (m *Manager) Render(ctx context.Context, w http.ResponseWriter, r *http.Request) (Result, error) {
	if m == nil {
		return Result{}, ErrNotFound
	}
	name, rest, err := Route(r.URL.Path)
	if err != nil {
		return Result{}, err
	}
	module, ok := m.Module(name)
	if !ok {
		return Result{}, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	if file, found := m.file(module, rest); found {
		return file, nil
	}
	return m.run(ctx, w, r, module)
}

// Route splits a request path into the module's name and what is left for the
// module to answer. A path with fewer than two segments after the prefix names
// no module.
func Route(urlPath string) (string, string, error) {
	trimmed := strings.TrimPrefix(path.Clean("/"+strings.TrimPrefix(urlPath, Prefix)), "/")
	vendor, rest, _ := strings.Cut(trimmed, "/")
	plugin, rest, _ := strings.Cut(rest, "/")
	name := vendor + "/" + plugin
	if _, _, err := ParseName(name); err != nil {
		return "", "", fmt.Errorf("%w: %s", ErrNotFound, urlPath)
	}
	return name, rest, nil
}

// file answers a path under the module's public directory, and reports whether
// one is there. A directory is not a file, and a path leaving the root is
// refused.
func (m *Manager) file(module *Module, rest string) (Result, bool) {
	if rest == "" {
		return Result{}, false
	}
	rel, err := Inside(module.Root, path.Join(Public, rest))
	if err != nil {
		return Result{}, false
	}
	full := filepath.Join(module.Root, rel)
	info, err := os.Stat(full)
	if err != nil || info.IsDir() {
		return Result{}, false
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return Result{}, false
	}
	header := http.Header{}
	if kind := mime.TypeByExtension(filepath.Ext(full)); kind != "" {
		header.Set("Content-Type", kind)
	}
	return Result{Body: data, Status: http.StatusOK, Header: header, File: true}, true
}

// run executes the module's index.php and returns what it echoed.
//
// A runtime comes from the module's pool and goes back to it, reset. The
// module is on the context the runtime reads, so the bindings know whose files
// Task::migrate and Task::dispatch may touch without taking it as an argument.
func (m *Manager) run(ctx context.Context, w http.ResponseWriter, r *http.Request, module *Module) (Result, error) {
	p, ok := m.poolOf(module.Name())
	if !ok {
		return Result{}, fmt.Errorf("%w: %s", ErrNotFound, module.Name())
	}
	if _, err := fs.Stat(os.DirFS(module.Root), Index); err != nil {
		return Result{}, fmt.Errorf("%s: %w", module.Name(), err)
	}

	ctx, cancel := context.WithTimeout(ctx, m.opts.Timeout)
	defer cancel()
	ctx = WithModule(ctx, module)

	rt, release := p.get()
	defer release()

	var out bytes.Buffer
	rt.PushOutput(&out)
	defer rt.EndRequest(rt.BeginRequest(ctx))
	rt.SetContext(ctx)

	request := runner.FromRequest(r.WithContext(ctx))
	defer request.Cleanup()
	serverVars(request, r, module)
	request.Register(rt)
	rt.AccountRequest(r, w)

	started := time.Now()
	prog, err := rt.LoadFile(Index)
	if err == nil {
		err = rt.Run(prog)
	}
	took := time.Since(started)

	if _, exited := runner.IsExit(err); err != nil && !exited {
		return Result{Took: took}, fmt.Errorf("%s: %w", module.Name(), err)
	}
	// A script that staged no status staged none: ResponseStatus answers zero
	// for that, and the page it rendered is a 200.
	status := request.StatusFor(nil)
	if status == 0 {
		status = http.StatusOK
	}
	return Result{
		Body:   out.Bytes(),
		Status: status,
		Header: request.ResponseHeaders(),
		Took:   took,
	}, nil
}

// serverVars fills the $_SERVER entries the request cannot derive on its own:
// where the module lives and which script answered.
//
// REQUEST_URI is left alone. It derives from the request, prefix and all, and
// that is the contract: a module reads the path the browser requested.
func serverVars(request runner.Context, r *http.Request, module *Module) {
	request.Server["DOCUMENT_ROOT"] = module.Root
	request.Server["SERVER_SOFTWARE"] = "task-ui"
	request.Server["SCRIPT_NAME"] = module.Href() + "/" + Index
	request.Server["PHP_SELF"] = module.Href() + "/" + Index
	request.Server["SCRIPT_FILENAME"] = filepath.Join(module.Root, Index)
	request.Server["TASK_UI_MODULE"] = module.Name()
	request.Server["TASK_UI_MODULE_BASE"] = module.Href()

	host := r.Host
	if name, port, err := net.SplitHostPort(host); err == nil {
		host = name
		request.Server["SERVER_PORT"] = port
	}
	request.Server["SERVER_NAME"] = strings.TrimSuffix(strings.ToLower(host), ".")
}
