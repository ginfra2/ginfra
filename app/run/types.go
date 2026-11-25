/*
Package run
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Orchestration runner types and values.
*/
package run

import (
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"os/exec"
	"strings"
)

// #####################################################################################################################
// # Build in variables that are automatically available to all scripts.

const RunBuiltInIp = "GINFRA_BUILTIN_IP"
const RunBuiltInDockerHost = "GINFRA_DOCKER_HOST"
const RunBuiltInLocal = "LOCAL"

// #####################################################################################################################
// # Tokens and notions

const (
	TOKEN_CAPTURE      = "capture"
	TOKEN_CMD          = "cmd"
	TOKEN_CODEEXPECT   = "codeexpect"
	TOKEN_COMMENT      = "comment"
	TOKEN_CONFIG       = "config"
	TOKEN_DELAY        = "delay"
	TOKEN_DEPLOYMENT   = "deployment"
	TOKEN_DETACH       = "detach"
	TOKEN_EXTRACT      = "extract"
	TOKEN_FAILOK       = "failok"
	TOKEN_GINFRA       = "ginfra"
	TOKEN_LOG          = "log"
	TOKEN_META         = "meta"
	TOKEN_NOSCAN       = "noscan"
	TOKEN_OUTPUT       = "output"
	TOKEN_PING         = "ping"
	TOKEN_PLATFORM     = "platform"
	TOKEN_POST         = "post"
	TOKEN_REGEX        = "regex"
	TOKEN_REMOVE       = "remove"
	TOKEN_REPEAT_DELAY = "repeatdelay"
	TOKEN_REPEAT       = "repeat"
	TOKEN_SCAN         = "scan"
	TOKEN_SOURCE       = "source"
	TOKEN_SUCCESS      = "success"
	TOKEN_TARGET       = "target"
	TOKEN_TERMINATE    = "terminate"
	TOKEN_UNTIL        = "until"
	TOKEN_URL          = "url"
	TOKEN_WORKFLOW     = "workflow"
	TOKEN_WORKINGDIR   = "workingdir"

	CMD_ANNOTATION_CHAR = "." // Deprecated for now, but I plan on reviving it.

	CONFIG_NV_SEP_CHAR = "="
)

// #####################################################################################################################
// # GENERAL ITEMS

// ScriptContext represents the context for executing scripts, holding configurations, working directory, and state details.
type ScriptContext struct {
	Script     *Script
	WorkingDir string
	Detached   []*GinfraInstance
	Configs    map[string]interface{}
	HostedSeen bool

	ProcessConfig  config.ProcessConfig
	ConfigProvider config.ConfigProvider
}

type Platform int

const (
	PAny Platform = iota
	PNix
	PWindows // Technically only windows matters to the logic.
)

var (
	platformMap = map[string]Platform{
		"any":     PAny,
		"nix":     PNix,
		"win":     PWindows,
		"windows": PWindows,
	}
)

func getPlatform(t string) (Platform, error) {
	var err error
	p, ok := platformMap[strings.ToLower(t)]
	if !ok {
		err = base.NewGinfraErrorA("Unknown platform name.", base.LM_NAME, t)
	}
	return p, err
}

type Cmd string

type Script struct {
	Meta       Meta
	Workflow   []*WorkflowItem
	SourceFile string
}

type Meta struct {
	WorkingDir *string
	Deployment *string
	Terminate  bool
	Config     *string
}

type WorkflowItem struct {
	Command *Command
	Ginfra  *Ginfra
	Log     *string
}

// Runnable TODO I will probably make Command detachable eventually.
type Runnable interface {
	IsKilled() bool
	SetKilled(k bool)
	SetCommand(ecmd *exec.Cmd)
	GetCommand() *exec.Cmd
}

func prepareS(sctx *ScriptContext, val string) (string, error) {
	s, err := config.ProcessVarConfigString(sctx.ConfigProvider, sctx.ProcessConfig, sctx.Configs, val)
	return s, err
}

func prepareI(sctx *ScriptContext, i any) (any, error) {
	switch val := i.(type) {
	case Cmd:
		sd := string(val)
		s, err := config.ProcessVarConfigString(sctx.ConfigProvider, sctx.ProcessConfig, sctx.Configs, sd)
		return Cmd(s), err

	case WorkingDir:
		s, err := config.ProcessVarConfigString(sctx.ConfigProvider, sctx.ProcessConfig, sctx.Configs, val.Path)
		return WorkingDir{Path: s}, err

	case Remove:
		var (
			r   Remove
			s   string
			err error
		)
		for _, ri := range val.Paths {
			s, err = config.ProcessVarConfigString(sctx.ConfigProvider, sctx.ProcessConfig, sctx.Configs, ri)
			if err != nil {
				return r, err
			}
			r.Paths = append(r.Paths, s)
		}
		return r, err

	case Output:
		var (
			ro  Output
			s   string
			err error
		)
		for _, di := range val.D {
			switch dival := di.(type) {
			case Delay:
				ro.D = append(ro.D, dival) // It's an int

			case Until:
				ro.D = append(ro.D, dival) // It's an int

			case Source:
				sd := string(dival)
				s, err = config.ProcessVarConfigString(sctx.ConfigProvider, sctx.ProcessConfig, sctx.Configs, sd)
				if err != nil {
					return ro, err
				}
				ro.D = append(ro.D, Source(s))

			case Scan:
				sd := string(dival)
				s, err = config.ProcessVarConfigString(sctx.ConfigProvider, sctx.ProcessConfig, sctx.Configs, sd)
				if err != nil {
					return ro, err
				}
				ro.D = append(ro.D, Scan(s))

			case NoScan:
				sd := string(dival)
				s, err = config.ProcessVarConfigString(sctx.ConfigProvider, sctx.ProcessConfig, sctx.Configs, sd)
				if err != nil {
					return ro, err
				}
				ro.D = append(ro.D, NoScan(s))

			case Extract:
				e := dival
				e.Target, err = config.ProcessVarConfigString(sctx.ConfigProvider, sctx.ProcessConfig, sctx.Configs, e.Target)
				if err != nil {
					return ro, err
				}
				e.Regex, err = config.ProcessVarConfigString(sctx.ConfigProvider, sctx.ProcessConfig, sctx.Configs, e.Regex)
				if err != nil {
					return ro, err
				}
				ro.D = append(ro.D, e)

			case Capture:
				c := dival
				c.Target, err = config.ProcessVarConfigString(sctx.ConfigProvider, sctx.ProcessConfig, sctx.Configs, c.Target)
				if err != nil {
					return ro, err
				}
				ro.D = append(ro.D, c)

			default:
				panic("BUG")
			}
		}
		return ro, err

	case Ping:
		p := val
		s, err := config.ProcessVarConfigString(sctx.ConfigProvider, sctx.ProcessConfig, sctx.Configs, p.Url)
		if err != nil {
			return p, err
		}
		p.Url = s
		return p, err

	default:
		panic("BUG: Unknown directive.  This should have been caught by the parser.")
	}
}

// Prepare processes input directives, applying variable replacements based on the script context and common directives.
// sctx is the script execution context containing configurations and process details.
// d is the slice of directives to be prepared.
// c represents the common directives providing structure and success handler details.
// Returns a processed slice of directives (rd) and an error if any issue occurs.
func Prepare(sctx *ScriptContext, d []interface{}, c *CommonDirectives) (rd []interface{}, err error) {
	rd = make([]interface{}, 0)

	var ii any
	for _, i := range d {
		if ii, err = prepareI(sctx, i); err != nil {
			return
		} else {
			rd = append(rd, ii)
		}

	}

	if c.Success != nil {
		var cs string
		if cs, err = common.SimpleReplacer(*c.Success, sctx.Configs, false); err != nil {
			return
		}
		c.Success = &cs
	}
	
	return
}

// #####################################################################################################################
// # DIRECTIVES - they tell the engine to do something in the given context.

type CommonDirectives struct {
	Success     *string
	Post        bool
	Repeat      Repeat
	RepeatDelay int
}

// Command runs a system command line with the given directives.  Directives that specify action are executed in the
// order given in the struct.
type Command struct {
	// Directives
	D []interface{}
	C CommonDirectives

	// Unique
	Platform Platform `json:"platform,omitempty"`
	FailOk   bool     `json:"fail_ok,omitempty"`

	// Meta
	killed bool      `json:"killed,omitempty"`
	ecmd   *exec.Cmd `json:"ecmd,omitempty"`
}

func (c *Command) IsKilled() bool {
	return c.killed
}

func (c *Command) SetKilled(k bool) {
	c.killed = k
}

func (c *Command) SetCommand(ecmd *exec.Cmd) {
	c.ecmd = ecmd
}

func (c *Command) GetCommand() *exec.Cmd {
	return c.ecmd
}

func (c *Command) Prepare(sctx *ScriptContext) error {
	var (
		err error
	)
	if c.D, err = Prepare(sctx, c.D, &c.C); err != nil {
		return err
	}
	return nil
}

type GinfraInstance struct {
	// Meta
	Killed bool
	Ecmd   *exec.Cmd
}

// Ginfra runs a Ginfra application command with the containing directives.
type Ginfra struct {
	// Directives
	D []interface{}
	C CommonDirectives

	// Unique
	Detach bool
}

func (c *GinfraInstance) IsKilled() bool {
	return c.Killed
}

func (c *GinfraInstance) SetKilled(k bool) {
	c.Killed = k
}

func (c *GinfraInstance) SetCommand(ecmd *exec.Cmd) {
	c.Ecmd = ecmd
}

func (c *GinfraInstance) GetCommand() *exec.Cmd {
	return c.Ecmd
}

func (c *Ginfra) Prepare(sctx *ScriptContext) error {
	var (
		err error
	)
	if c.D, err = Prepare(sctx, c.D, &c.C); err != nil {
		return err
	}
	return nil
}

// Extract the first matches regex and put it in the variable named in Target.
type Extract struct {
	Regex  string
	Target string
}

// Capture all output and put it in the variable named in Target.  Making it s struct since it will eventually get more
// features.
type Capture struct {
	Target string
}

// Output check.
type Output struct {
	// Only one
	Terminate bool
	Repeat    Repeat

	// List
	D []interface{}
}

// Remove the listed Paths, both files and directories.
type Remove struct {
	Paths []string
}

// Repeat the containing directive int many times.
type Repeat int

// Delay the containing directive int many milliseconds.
type Delay int

// Until will continue the containing directive until in int many milliseconds have passed.
type Until int

// Source file.
type Source string

// Scan for the given string.
type Scan string

// NoScan will scan for the given string but will fail if it is actually found.
type NoScan string

// WorkingDir sets the working directory.
type WorkingDir struct {
	Path string
}

// Ping defines configuration settings for performing a ping operation to a specified URL.
type Ping struct {
	Delay      Delay
	Until      int
	Url        string
	CodeExpect int
	Repeat     Repeat
}
