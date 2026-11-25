/*
Package run
Copyright (c) 2023, 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Orchestration runner.  Run() will parse the script file and run it.  RunScript() will run a parsed run.Script object.
This is not intended for heavy duty production usage.
*/
package run

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// #####################################################################################################################
// #

const ScanPollInterval = 50

// NewScriptContext initializes and returns a new instance of ScriptContext with default configuration values.
func NewScriptContext(path string) *ScriptContext {
	var s ScriptContext
	s.Configs = make(map[string]interface{})

	// == BUILTINs ============================
	oip := common.GetOutboundIP()
	if oip == "" {
		oip = "0.0.0.0" // TODO loopback instead?
	}

	s.Configs[RunBuiltInIp] = oip
	s.Configs[RunBuiltInDockerHost] = oip // TODO make configurable.
	s.Configs[base.EnvGinfraHome] = base.GetGinfraHome()
	s.Configs[base.EnvGinfraTestHome] = base.GetGinfraTmpTestDir()
	s.Configs[RunBuiltInLocal] = filepath.Dir(path)

	return &s
}

func execCheckPlatform(p Platform) bool {
	// TODO just brute force it now.
	if p == PWindows {
		hp := runtime.GOOS // If I don't do this the compiler optimizes this entire section out.
		if hp != "windows" {
			return false
		}
	} else {
		hp := runtime.GOOS
		if hp == "windows" {
			return false
		}
	}
	return true
}

func doOutputSource(out string) (string, error) {
	w, _ := os.Getwd()
	b, err := os.ReadFile(out)
	if err != nil {
		// try again with a full path.
		b, err = os.ReadFile(filepath.Join(w, out))
		if err != nil {
			wd, _ := os.Getwd()
			return "", base.NewGinfraErrorChildA("OUTPUT SOURCE not a valid path.", err, base.LM_FILEPATH,
				out, base.LM_TARGET, wd)
		}
	}
	return string(b), nil
}

func itemsProcessOutput(gctx *app.GContext, sctx *ScriptContext, o *Output, out string) error {
	var (
		err   error
		until *time.Time
		csrc  string
	)

	for _, i := range o.D {
		// clear until?
		if until != nil {
			if time.Now().After(*until) {
				until = nil
			}
		}

		switch oi := i.(type) {

		case Delay:
			gctx.GetLogger().Infow("Delay in milliseconds.", zap.String(base.LM_TIME, strconv.Itoa(int(oi))))
			time.Sleep(time.Duration(oi) * time.Millisecond)

		case Until:
			gctx.GetLogger().Infow("Until in milliseconds.", zap.String(base.LM_TIME, strconv.Itoa(int(oi))))
			u := time.Now().Add(time.Millisecond * time.Duration(oi))
			until = &u

		case Source:
			csrc = string(oi)
			if until == nil {
				out, err = doOutputSource(string(oi))
			} else {
				out, err = doOutputSource(string(oi))
				if err != nil {
					for !time.Now().After(*until) {
						out, err = doOutputSource(string(oi))
						if err == nil {
							break
						}
						time.Sleep(ScanPollInterval * time.Millisecond)
					}
				}
				if err != nil {
					return err
				}
			}

		case Scan:
			if until == nil {
				if strings.Index(out, string(oi)) < 0 {
					return base.NewGinfraErrorA("Cmd OUTPUT SCAN failed.", base.LM_OUTPUT, out, base.LM_SCAN,
						oi)
				}
			} else {
				if csrc == "" {
					// I might be able to fix this, but it'll be complicated.
					return base.NewGinfraErrorA("Cmd OUTPUT SCAN on UNTIL does not work unless SOURCE is set.  No additional output from the command line can be read.",
						base.LM_OUTPUT, out, base.LM_SCAN, oi)
				}

				sc := false
				for !time.Now().After(*until) {
					if strings.Index(out, string(oi)) >= 0 {
						sc = true
						break
					}
					time.Sleep(ScanPollInterval * time.Millisecond)
					out, _ = doOutputSource(csrc)
				}
				if !sc {
					return base.NewGinfraErrorA("Cmd OUTPUT SCAN failed after UNTIL timeout.", base.LM_OUTPUT, out, base.LM_SCAN,
						oi)
				}
			}
			gctx.GetLogger().Infow("SCAN successful.")

		case NoScan:
			if until == nil {
				if strings.Index(out, string(oi)) >= 0 {
					return base.NewGinfraErrorA("Cmd OUTPUT NOSCAN failed.", base.LM_OUTPUT, out, base.LM_SCAN,
						oi)
				}
			} else {
				if csrc == "" {
					// I might be able to fix this, but it'll be complicated.
					return base.NewGinfraErrorA("Cmd OUTPUT NOSCAN on UNTIL does not work unless SOURCE is set.  No additional output from the command line can be read.",
						base.LM_OUTPUT, out, base.LM_SCAN, oi)
				}

				sc := false
				for !time.Now().After(*until) {
					if strings.Index(out, string(oi)) < 0 {
						sc = true
						break
					}
					time.Sleep(ScanPollInterval * time.Millisecond)
					out, _ = doOutputSource(csrc)
				}
				if !sc {
					return base.NewGinfraErrorA("Cmd OUTPUT NOSCAN failed after UNTIL timeout.", base.LM_OUTPUT, out, base.LM_SCAN,
						oi)
				}
			}
			gctx.GetLogger().Infow("NoScan successful.")

		case Extract:
			m, err := regexp.Compile(oi.Regex)
			if err != nil {
				return base.NewGinfraErrorChildA("Bad RegEx in EXTRACT.", err, base.LM_TEXT, oi.Regex)
			}
			s := m.FindString(out)
			if s != "" {
				sctx.Configs[oi.Target] = s
			}

		case Capture:
			sctx.Configs[oi.Target] = out

		} // end switch

	}

	return err
}

func processOutput(gctx *app.GContext, sctx *ScriptContext, o *Output, out string) error {
	var (
		err error
	)

	if o != nil {
		for idx := 0; idx <= int(o.Repeat); idx++ {
			err = itemsProcessOutput(gctx, sctx, o, out)
			if err == nil {
				return nil
			}
			if idx < int(o.Repeat) {
				gctx.GetLogger().Infow(fmt.Sprintf("Output failed.  Trying again.  %d of %d", idx+1, int(o.Repeat)+1))
			}
		}
	}
	return err
}

const UNTIL_ITERATION_DELAY = 200

func itemProcessPing(gctx *app.GContext, p *Ping) error {
	if p.Delay > 0 {
		gctx.GetLogger().Infow("Delay in milliseconds.", zap.String(base.LM_TIME, strconv.Itoa(int(p.Delay))))
		time.Sleep(time.Duration(p.Delay) * time.Millisecond)
	}

	var (
		done  bool
		didto bool
		didct bool
		res   *http.Response
		err   error
		url   = common.CleanUrl(p.Url)
	)

	tick := time.NewTicker(time.Millisecond * UNTIL_ITERATION_DELAY)
	defer tick.Stop()
	to := time.NewTimer(time.Millisecond * time.Duration(p.Until))
	defer to.Stop()
	for {
		if done {
			break
		}
		select {
		case <-to.C:
			// timeout
			didto = true
			done = true

		case <-tick.C:
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}
			if res, err = client.Get(url); err == nil {
				if res.StatusCode == p.CodeExpect {
					// Done and ok
					done = true
				} else {
					didct = true
				}
			}
		}
	}

	if didto {
		if didct {
			err = base.NewGinfraErrorA("PING did not get the expected status code in time.", base.LM_EXPECTED,
				p.CodeExpect, base.LM_ACTUAL, res.StatusCode, base.LM_URL, url)
		} else {
			err = base.NewGinfraErrorChildA("PING could not connect.", err, base.LM_URL, url)
		}
	} else {
		gctx.GetLogger().Infow("Ping succeeded.", zap.String(base.LM_URL, url))
	}

	return err
}

func processPing(gctx *app.GContext, p *Ping) error {
	var (
		err error
	)

	if p != nil {
		for idx := 0; idx <= int(p.Repeat); idx++ {
			err = itemProcessPing(gctx, p)
			if err == nil {
				return nil
			}
			if idx < int(p.Repeat) {
				gctx.GetLogger().Infow(fmt.Sprintf("Ping failed.  Trying again.  %d of %d", idx+1, int(p.Repeat)+1))
			}
		}
	}
	return err
}

// removeGlob will quietly swallow any errors.
func removeGlob(g string) {
	fi, _ := filepath.Glob(g)
	for _, f := range fi {
		_ = os.Remove(f)
	}
}

func doExecWorkflowCmd(gctx *app.GContext, c *Command, cmd string, failOk bool) (*exec.Cmd, string, error) {
	// Execute.
	var (
		cmdstr string
		cexec  *exec.Cmd
		o      []byte
		err    error
	)

	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil, "", base.NewGinfraError("No command present.")
	}
	sp := strings.Split(cmd, " ")
	cmdstr = sp[0]
	if len(sp) > 1 {
		sp = sp[1:]
	} else {
		sp = []string{}
	}

	for idx := 0; idx <= int(c.C.Repeat); idx++ {
		if runtime.GOOS == "windows" {
			sp = append([]string{"/c", cmdstr}, sp...)
			cexec = exec.Command("cmd.exe", sp...)
		} else {
			cexec = exec.Command(cmdstr, sp...)
		}

		if idx == 0 {
			gctx.GetLogger().Infow("Starting command.", zap.String(base.LM_COMMAND, cexec.String()))
		}

		if o, err = cexec.CombinedOutput(); err != nil && failOk == false {
			err = base.NewGinfraErrorChildA("Cmd failed.", err, base.LM_COMMAND, cmd,
				base.LM_ERROR, string(o))
		} else {
			if c.C.Post {
				gctx.GetLogger().Infow("Command output", zap.String(base.LM_OUTPUT, string(o)))
			}
			break
		}

		if idx < int(c.C.Repeat) {
			gctx.GetLogger().Infow(fmt.Sprintf("Ginfra CMD failed.  Trying again.  %d of %d", idx+2, int(c.C.Repeat)+1))
			time.Sleep(time.Duration(c.C.RepeatDelay) * time.Millisecond)
		}
	}

	return cexec, string(o), err
}

func execWorkflowCmd(gctx *app.GContext, sctx *ScriptContext, c *Command) error {
	var (
		err  error
		fail bool
	)

	// Platform check
	if c.Platform != PAny {
		if execCheckPlatform(c.Platform) == false {
			return nil // Ignore the c.
		}
	}

	// Restore CWD - assume for now
	defer func() {
		_ = os.Chdir(sctx.WorkingDir)
	}()

	for _, i := range c.D {
		switch ic := i.(type) {

		case Cmd:
			ip, lerr := prepareI(sctx, ic)
			if lerr != nil {
				return lerr
			}

			cexec, o, lerr := doExecWorkflowCmd(gctx, c, string(ip.(Cmd)), c.FailOk)
			if lerr != nil {
				if c.FailOk {
					gctx.GetLogger().Warn("Command failed, but allowed by FailOk.")
				} else {
					return lerr
				}
			}
			c.SetCommand(cexec)
			if c.C.Post {
				gctx.GetLogger().Infow("Command output", zap.String(base.LM_OUTPUT, o))
			}
			gctx.MostRecentOutput = o

		case WorkingDir:
			err = os.Chdir(ic.Path)
			if err == nil {
				gctx.GetLogger().Errorw("WorkingDir changed.", zap.String(base.LM_PATH, ic.Path))
			} else {
				return base.NewGinfraErrorChildA("Bad "+TOKEN_WORKINGDIR+" set in "+TOKEN_CMD+".", err,
					base.LM_FILEPATH, ic.Path)
			}

		case Remove:
			ip, lerr := prepareI(sctx, ic)
			if lerr != nil {
				return lerr
			}
			for _, ri := range ip.(Remove).Paths {
				removeGlob(ri)
			}

		case Output:
			ip, lerr := prepareI(sctx, ic)
			if lerr != nil {
				return lerr
			}
			io := ip.(Output)
			err = processOutput(gctx, sctx, &io, gctx.MostRecentOutput)
			if err == nil {
				gctx.GetLogger().Errorw("Output succeeded.")
			} else {
				fail = true
				gctx.GetLogger().Errorw("Output failed.", zap.String(base.LM_CAUSE, err.Error()))
			}

		case Ping:
			ip, lerr := prepareI(sctx, ic)
			if lerr != nil {
				return lerr
			}
			io := ip.(Ping)
			err = processPing(gctx, &io)
			if err != nil {
				fail = true
				gctx.GetLogger().Errorw("Ping failed.", zap.String(base.LM_CAUSE, err.Error()))
			}

		}
	}

	// Success log?
	if c.C.Success != nil && fail == false {
		if slog, err := prepareS(sctx, *c.C.Success); err == nil {
			gctx.GetLogger().Info(slog)
		} else {
			return base.NewGinfraErrorChild("Could not process success message for logging", err)
		}
	}

	return err
}

func doExecWorkflowGinfraCommand(gctx *app.GContext, sctx *ScriptContext, g *Ginfra, gcmd string) (*GinfraInstance, string, error) {

	var (
		err   error
		ginst GinfraInstance
		o     []byte
	)

	// Get command elements.  Is deployment set?
	ce := strings.Split(gcmd, " ")
	if sctx.Script.Meta.Deployment != nil {
		p := []string{"--deployment", *sctx.Script.Meta.Deployment}
		ce = append(p, ce...)
	}

	// Execute.
	// From here and below err should be returned from the last statement, so that this function has a chance to
	// restore state and note any processes that will need killing.
	// TODO was thinking maybe take over cmd.RootCmd, but lets' just be cheap for now.  The problem with taking over
	// rootcmd, is there is only one instance per application run and deconflicting it might be a nightmare.

	if g.Detach {
		cmd := exec.Command("ginfra", ce...)
		DecorateCommand(gctx, cmd)
		ginst.Ecmd = cmd

		if g.C.Repeat > 0 {
			gctx.GetLogger().Warn("REPEAT is ignored because DETACH is set.")
		}

		// TODO make sure these are killed on exit.  Windows has a bad habit of ghosting them.
		go func(g *GinfraInstance, post bool) {
			gctx.GetLogger().Infow("Starting detached.", zap.String(base.LM_COMMAND, cmd.String()))
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			err = cmd.Run()
			// Eat the error if it was caused by the kill.
			if err != nil && g.IsKilled() == false {
				gctx.GetLogger().Error("Could not run command.", zap.Error(err), zap.String(base.LM_COMMAND,
					cmd.String()), zap.String(base.LM_STDOUT, stdout.String()), zap.String(base.LM_STDERR, stderr.String()))
			} else {
				err = nil
				if post {
					gctx.GetLogger().Infow("Command output.", zap.String(base.LM_STDOUT, stdout.String()),
						zap.String(base.LM_STDERR, stderr.String()))
				}
			}

		}(&ginst, g.C.Post)

		// TODO Give it time for trivial error.  I really hate using times like this, but the effort to get around it
		// is way more than I can justify right now.
		time.Sleep(15 * time.Millisecond)

	} else {

		for idx := 0; idx <= int(g.C.Repeat); idx++ {
			cmd := exec.Command("ginfra", ce...)
			DecorateCommand(gctx, cmd)
			ginst.Ecmd = cmd

			if idx == 0 {
				gctx.GetLogger().Infow("Starting ginfra command.", zap.String(base.LM_COMMAND, cmd.String()))
			}

			if o, err = cmd.CombinedOutput(); err != nil {
				err = base.NewGinfraErrorChildA("Cmd failed.", err, base.LM_COMMAND, gcmd,
					base.LM_ERROR, string(o))
			} else {
				if g.C.Post {
					gctx.GetLogger().Infow("Command output", zap.String(base.LM_OUTPUT, string(o)))
				}
				break
			}

			if idx < int(g.C.Repeat) {
				gctx.GetLogger().Infow(fmt.Sprintf("Ginfra CMD failed.  Trying again.  %d of %d", idx+2, int(g.C.Repeat)+1))
				time.Sleep(time.Duration(g.C.RepeatDelay) * time.Millisecond)
			}
		}
	}

	return &ginst, string(o), err
}

type workflowGinfraDispatchContext struct {
	local     []*GinfraInstance
	fail      bool
	terminate bool
	err       error
}

func execWorkflowGinfraDispatch(gctx *app.GContext, sctx *ScriptContext, g *Ginfra) workflowGinfraDispatchContext {
	var (
		r workflowGinfraDispatchContext
	)

	for _, i := range g.D {
		switch ic := i.(type) {

		case Cmd:
			ip, err := prepareI(sctx, ic)
			if err != nil {
				r.err = err
				return r
			}
			li, o, err := doExecWorkflowGinfraCommand(gctx, sctx, g, string(ip.(Cmd)))
			if err != nil {
				r.err = err
				return r
			}
			r.local = append(r.local, li)
			gctx.MostRecentOutput = o

		case WorkingDir:
			err := os.Chdir(ic.Path)
			if err == nil {
				gctx.GetLogger().Errorw("WorkingDir changed.", zap.String(base.LM_PATH, ic.Path))
			} else {
				r.err = err
				return r
			}

		case Remove:
			ip, err := prepareI(sctx, ic)
			if err != nil {
				r.err = err
				return r
			}
			for _, ri := range ip.(Remove).Paths {
				removeGlob(ri)
			}

		case Output:
			ip, err := prepareI(sctx, ic)
			if err != nil {
				r.err = err
				return r
			}
			io := ip.(Output)
			err = processOutput(gctx, sctx, &io, gctx.MostRecentOutput)
			if err == nil {
				gctx.GetLogger().Errorw("Output succeeded.")
			} else {
				r.err = err
				r.fail = true
				gctx.GetLogger().Errorw("Output failed.", zap.String(base.LM_CAUSE, err.Error()))
			}
			if ic.Terminate {
				r.terminate = true
			}

		case Ping:
			ip, err := prepareI(sctx, ic)
			if err != nil {
				r.err = err
				return r
			}
			io := ip.(Ping)
			err = processPing(gctx, &io)
			if err != nil {
				r.err = err
				r.fail = true
				gctx.GetLogger().Errorw("Ping failed.", zap.String(base.LM_CAUSE, err.Error()))
			}
		}
	}

	return r
}

func execWorkflowGinfra(gctx *app.GContext, sctx *ScriptContext, g *Ginfra) error {

	r := execWorkflowGinfraDispatch(gctx, sctx, g)

	// Make sure the process is actually terminated.  Windows sometimes ghosts when there is in error in
	// cmd.Run().
	if g.Detach {
		if r.terminate {
			// Now that output is done, terminate it.
			for _, ti := range r.local {
				gctx.GetLogger().Info("Terminating as ordered.")
				ti.SetKilled(true)
				TerminateCommand(gctx, ti.GetCommand())
			}

		} else {
			// Leave detached
			sctx.Detached = append(sctx.Detached, r.local...)
		}
	}

	// Success log?
	if g.C.Success != nil && r.fail == false {
		if slog, err := prepareS(sctx, *g.C.Success); err == nil {
			gctx.GetLogger().Info(slog)
		} else {
			return base.NewGinfraErrorChild("Could not process success message for logging", err)
		}
	}

	// Restore CWD - just assume it was done for now.
	_ = os.Chdir(sctx.WorkingDir)

	// See if a host file has appeared.
	checkHostFile(gctx, sctx)

	return r.err
}

func execWorkflow(gctx *app.GContext, sctx *ScriptContext, wf *WorkflowItem) error {
	var (
		s   string
		err error
	)

	if wf.Command != nil {
		// if err = wf.Command.Prepare(sctx); err == nil {
		err = execWorkflowCmd(gctx, sctx, wf.Command)
		//}
	}
	if err == nil && wf.Ginfra != nil {
		//if err = wf.Ginfra.Prepare(sctx); err == nil {
		err = execWorkflowGinfra(gctx, sctx, wf.Ginfra)
		//}
	}
	if err == nil && wf.Log != nil {
		s, err = prepareS(sctx, *wf.Log)
		if err == nil {
			gctx.GetLogger().Info(s)
		} else {
			gctx.GetLogger().Errorw("Poorly formatted LOG.", zap.String(base.LM_CAUSE, err.Error()))
		}
	}

	return err
}

func handleDetached(gctx *app.GContext, sctx *ScriptContext, script *Script, err error) {
	if len(sctx.Detached) > 0 {
		if script.Meta.Terminate || err != nil {
			gctx.GetLogger().Info("Terminating detached.")
			for _, c := range sctx.Detached {
				c.SetKilled(true)
				TerminateCommand(gctx, c.GetCommand())
			}
		} else {
			gctx.GetLogger().Info("There are detached processes.  They will run until you kill this.")
			WaitKilled(gctx)
			terminateDeferred(gctx, sctx)
		}
	}
}

// RunScript executes a given script within the provided application context, handling its workflow and configurations.
// Returns an error if any issue occurs during execution.
func RunScript(gctx *app.GContext, script *Script) error {

	var (
		sctx *ScriptContext
		err  error
	)
	sctx = NewScriptContext(script.SourceFile)
	sctx.Script = script
	checkHostFile(gctx, sctx)

	// Make sure the detached are properly terminated instead of ghosted.  The latter seems to happen sometimes
	// on windows when the app does not exit gracefully.
	defer func() {
		if r := recover(); r != nil {
			terminateDeferred(gctx, sctx)
			panic(r)
		}
	}()

	// Meta
	if script.Meta.WorkingDir == nil {
		sctx.WorkingDir, _ = os.Getwd()

	} else {
		sctx.WorkingDir = *script.Meta.WorkingDir
		if err = os.Chdir(sctx.WorkingDir); err != nil {
			return base.NewGinfraErrorChildA("Bad WorkingDir set in Meta.", err, base.LM_FILEPATH,
				sctx.WorkingDir)
		}
	}

	// Relative to where the script is.
	if script.Meta.Config != nil {
		sdir := filepath.Dir(script.SourceFile)

		// Some weird stuff happening on windows so do this.
		if sdir == "." {
			sdir, _ = os.Getwd()
		}
		p := path.Join(sdir, *script.Meta.Config)

		file, errs := os.Open(p)
		if errs != nil {
			return base.NewGinfraErrorChildA("Bad Config file specified in Meta.  Could not open.", errs,
				base.LM_FILEPATH, *script.Meta.Config)
		}
		defer file.Close()

		s := bufio.NewScanner(file)
		// optionally, resize scanner's capacity for lines over 64K, see next example
		for s.Scan() {
			l := s.Text()
			if len(l) > 0 {
				sa := strings.Split(l, CONFIG_NV_SEP_CHAR)
				if len(sa) != 2 {
					return base.NewGinfraErrorChildA("Bad Config file specified in Meta.  Invalid entry.  They must be name value pairs separated by a "+CONFIG_NV_SEP_CHAR,
						errs, base.LM_FILEPATH, *script.Meta.Config)
				}

				sctx.Configs[strings.TrimSpace(sa[0])] = strings.TrimSpace(sa[1])
			}
		}

		if errs = s.Err(); errs != nil {
			return base.NewGinfraErrorChildA("Bad Config file specified in Meta.  Error while reading.", errs,
				base.LM_FILEPATH, *script.Meta.Config)
		}
	}

	// Setup traps
	SetupINTTrap(gctx, sctx)

	// Workflow
	for _, item := range script.Workflow {
		err = execWorkflow(gctx, sctx, item)
		if err != nil {
			break
		}
	}

	handleDetached(gctx, sctx, script, err)

	return err
}

// Run executes a script defined in the specified YAML file within the given application context.
// Returns an error if the script cannot be loaded or executed.
func Run(gctx *app.GContext, path string) error {

	// Just supporting yaml for now.
	s, err := LoadAndParseYaml(path)
	if err != nil {
		return err
	}

	return RunScript(gctx, s)
}
