/*
Package run
Copyright (c) 2023, 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Orchestration runner yaml script parser.  The script executor doesn't actually care about yaml.  This will convert yaml
into a script object.  You could create a script object any way you want, including compiling it from other syntax.

However, this parser does enforce certain rules for how a script object may be constructed, so as to prevent runtime
errors that might leave the system in a partial or inconsistent state.
*/
package run

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"gitlab.com/ginfra/ginfra/base"
	"os"
	"strconv"
	"strings"
)

// #####################################################################################################################
// # YAML PARSING

func guardString(user string, vv interface{}, errs []error) (string, []error) {
	var r string
	if vv != nil {
		if vs, ok := vv.(string); ok {
			r = vs
		} else {
			errs = append(errs, base.NewGinfraError("Malformed "+user+".  Expecting a string."))
		}
	} else {
		errs = append(errs, base.NewGinfraErrorA("Empty "+user+"."))
	}
	return r, errs
}

func guardTruth(user string, vv interface{}, errs []error) (bool, []error) {
	var r bool
	if vv != nil {
		switch vv.(type) {
		case string:
			r = base.Truth(vv.(string))
		case bool:
			r = vv.(bool)
		default:
			errs = append(errs, base.NewGinfraError("Malformed "+user+".  Expecting a string (with a bool value) or a boolean."))
		}
	} else {
		errs = append(errs, base.NewGinfraErrorA("Empty "+user+"."))
	}
	return r, errs
}

func guardInt(user string, v interface{}, errs []error) (int, []error) {
	var (
		r   int
		err error
	)
	if v != nil {
		switch vv := v.(type) {
		case string:
			r, err = strconv.Atoi(vv)
			if err != nil {
				errs = append(errs, base.NewGinfraErrorChildA("Bad value for "+user+".  Could not parse from string",
					err, base.LM_VALUE, vv))
			}
		case uint64:
			r = int(vv) // TODO serious precision problems.

		case int:
			r = vv

		default:
			errs = append(errs, base.NewGinfraError("Malformed "+user+".  Expecting a string (with a int value) or an integer."))
		}
	} else {
		errs = append(errs, base.NewGinfraErrorA("Empty "+user+"."))
	}
	return r, errs
}

func parseYamlMeta(s *Script, in any, errs []error) []error {
	if innn, ok := in.(map[string]interface{}); ok {
		for k, v := range innn {
			switch strings.ToLower(k) {
			case TOKEN_WORKINGDIR:
				var str string
				str, errs = guardString(TOKEN_META+":"+TOKEN_WORKINGDIR, v, errs)
				s.Meta.WorkingDir = &str

			case TOKEN_DEPLOYMENT:
				var str string
				str, errs = guardString(TOKEN_META+":"+TOKEN_DEPLOYMENT, v, errs)
				s.Meta.Deployment = &str

			case TOKEN_TERMINATE:
				s.Meta.Terminate, errs = guardTruth(TOKEN_META+":"+TOKEN_TERMINATE, v, errs)

			case TOKEN_COMMENT:
				break // NOP

			case TOKEN_CONFIG:
				var str string
				str, errs = guardString(TOKEN_META+":"+TOKEN_CONFIG, v, errs)
				s.Meta.Config = &str

			default:
				errs = append(errs, base.NewGinfraErrorA("Bad META section.  Unknown token.", base.LM_TOKEN, k))
			}
		}
	} else {
		errs = append(errs, base.NewGinfraErrorA("Top-level META section should be an object"))
	}
	return errs // TODO I don't think I'm thrashing memory here, but I should check.
}

func parseYamlPing(in interface{}, errs []error) (Ping, []error) {
	var p Ping

	if inn, ok := in.(map[string]interface{}); ok {
		for k, v := range inn {
			switch strings.ToLower(k) {
			case TOKEN_DELAY:
				var di int
				di, errs = guardInt(TOKEN_PING+":"+TOKEN_DELAY, v, errs)
				p.Delay = Delay(di)
			case TOKEN_REPEAT:
				var ri int
				ri, errs = guardInt(TOKEN_PING+":"+TOKEN_REPEAT, v, errs)
				p.Repeat = Repeat(ri)
			case TOKEN_UNTIL:
				p.Until, errs = guardInt(TOKEN_PING+":"+TOKEN_UNTIL, v, errs)
			case TOKEN_URL:
				p.Url, errs = guardString(TOKEN_PING+":"+TOKEN_URL, v, errs)
			case TOKEN_CODEEXPECT:
				p.CodeExpect, errs = guardInt(TOKEN_PING+":"+TOKEN_CODEEXPECT, v, errs)
			case TOKEN_COMMENT:
				break // NOP
			default:
				errs = append(errs, base.NewGinfraErrorA("Unknown command PING directive.", base.LM_TOKEN, k))
			}
		}
	} else {
		errs = append(errs, base.NewGinfraError("PING must be a yaml object."))
	}
	return p, errs
}

func parseYamlExtract(in map[string]interface{}, errs []error) (Extract, []error) {
	var e Extract
	for k, v := range in {
		switch strings.ToLower(k) {
		case TOKEN_REGEX:
			e.Regex, errs = guardString(TOKEN_EXTRACT+":"+TOKEN_REGEX, v, errs)
		case TOKEN_TARGET:
			e.Target, errs = guardString(TOKEN_EXTRACT+":"+TOKEN_TARGET, v, errs)
		case TOKEN_COMMENT:
			break // NOP
		default:
			errs = append(errs, base.NewGinfraErrorA("Unknown command in EXTRACT directive.", base.LM_TOKEN, k))
		}
	}
	return e, errs
}

func parseYamlCapture(in map[string]interface{}, errs []error) (Capture, []error) {
	var c Capture
	for k, v := range in {
		switch strings.ToLower(k) {
		case TOKEN_TARGET:
			c.Target, errs = guardString(TOKEN_EXTRACT+":"+TOKEN_TARGET, v, errs)
		case TOKEN_COMMENT:
			break // NOP
		default:
			errs = append(errs, base.NewGinfraErrorA("Unknown command in CAPTURE directive.", base.LM_TOKEN, k))
		}
	}
	return c, errs
}

func parseYamlOutput(in interface{}, errs []error) (Output, []error) {
	var (
		t []interface{}
		o Output
	)

	// If it is a map, turn it into an array,
	switch ct := in.(type) {
	case []interface{}:
		t = ct

	case map[string]interface{}:
		for ik, iv := range ct {
			t = append(t, map[string]interface{}{ik: iv})
		}

	default:
		errs = append(errs, base.NewGinfraErrorA("Unknown type for OUTPUT.  It must be an object of list of object."))
		return o, errs
	}

	for _, i := range t {
		if iv, ok := i.(map[string]interface{}); ok {
			for kk, vv := range iv {

				switch strings.ToLower(kk) {
				case TOKEN_TERMINATE:
					o.Terminate, errs = guardTruth(TOKEN_OUTPUT+":"+TOKEN_TERMINATE, vv, errs)

				case TOKEN_SCAN:
					var str string
					str, errs = guardString(TOKEN_OUTPUT+":"+TOKEN_SCAN, vv, errs)
					o.D = append(o.D, Scan(str))

				case TOKEN_NOSCAN:
					var str string
					str, errs = guardString(TOKEN_OUTPUT+":"+TOKEN_NOSCAN, vv, errs)
					o.D = append(o.D, NoScan(str))

				case TOKEN_SOURCE:
					var str string
					str, errs = guardString(TOKEN_OUTPUT+":"+TOKEN_SOURCE, vv, errs)
					o.D = append(o.D, Source(str))

				case TOKEN_UNTIL:
					var ui int
					ui, errs = guardInt(TOKEN_OUTPUT+":"+TOKEN_UNTIL, vv, errs)
					o.D = append(o.D, Until(ui))

				case TOKEN_DELAY:
					var di int
					di, errs = guardInt(TOKEN_OUTPUT+":"+TOKEN_DELAY, vv, errs)
					o.D = append(o.D, Delay(di))

				case TOKEN_REPEAT:
					var ri int
					ri, errs = guardInt(TOKEN_OUTPUT+":"+TOKEN_REPEAT, vv, errs)
					o.Repeat = Repeat(ri)

				case TOKEN_EXTRACT:
					var e Extract
					if vvv, okk := vv.(map[string]interface{}); okk {
						e, errs = parseYamlExtract(vvv, errs)
						o.D = append(o.D, e)
					} else {
						errs = append(errs, base.NewGinfraError("EXTRACT must be an object, not nil, array or other entity.  If the EXTRACT is in a list, be certain REGEX and TARGET items are double tabbed below it."))
					}

				case TOKEN_CAPTURE:
					var c Capture
					if vvv, okk := vv.(map[string]interface{}); okk {
						c, errs = parseYamlCapture(vvv, errs)
						o.D = append(o.D, c)
					} else {
						errs = append(errs, base.NewGinfraError("CAPTURE must be an object, not nil, array or other entity.  If the EXTRACT is in a list, be certain REGEX and TARGET items are double tabbed below it."))
					}

				case TOKEN_COMMENT:
					break // NOP

				default:
					errs = append(errs, base.NewGinfraErrorA("Unknown command in OUTPUT directive.", base.LM_TOKEN, kk))
				}
			}

		} else {
			errs = append(errs, base.NewGinfraErrorA("Bad type in OUTPUT list."))
			break
		}
	}

	return o, errs
}

func parseYamlPlatform(t string, c *Command, errs []error) []error {
	p, err := getPlatform(t)
	if err != nil {
		errs = append(errs, err)
	} else {
		c.Platform = p
	}
	return errs
}

func tokenRemove(vv interface{}, errs []error) (Remove, []error) {
	var r Remove
	if vv != nil {
		if ws, aok := vv.(string); aok {
			ss := strings.Split(ws, ",")
			for _, sss := range ss {
				r.Paths = append(r.Paths, sss)
			}
		} else {
			errs = append(errs, base.NewGinfraErrorA("Malformed REMOVE.  Expecting a string."))
		}
	} else {
		errs = append(errs, base.NewGinfraErrorA("Empty REMOVE."))
	}
	return r, errs
}

func tokenPing(vv interface{}, errs []error, hasCmd, cmdSeen bool) (Ping, []error) {
	var p Ping
	if vv != nil {
		if hasCmd && !cmdSeen {
			errs = append(errs, base.NewGinfraError("PING before CMD is not allowed."))
		} else {
			p, errs = parseYamlPing(vv, errs)
		}
	} else {
		errs = append(errs, base.NewGinfraErrorA("Empty PING."))
	}
	return p, errs
}

func tokenOutput(vv interface{}, errs []error, hasCmd, cmdSeen bool) (Output, []error) {
	var o Output
	if vv != nil {
		if hasCmd && !cmdSeen {
			errs = append(errs, base.NewGinfraError("OUTPUT before CMD is not allowed."))
		} else {
			o, errs = parseYamlOutput(vv, errs)
		}
	} else {
		errs = append(errs, base.NewGinfraErrorA("Empty OUTPUT."))
	}
	return o, errs
}

func parseYamlCmd(s *Script, in []interface{}, hasCmd bool, errs []error) []error {

	c := Command{
		Platform: PAny,
		D:        make([]interface{}, 0),
	}

	var cmdSeen bool

	for _, v := range in {

		if ivv, ok := v.(map[string]interface{}); ok {

			for kk, vv := range ivv {

				// Is it annotated?
				/*
					var ann string
					if strings.Index(strings.ToLower(k), CMD_ANNOTATION_CHAR) >= 0 {
						stok := strings.Split(k, CMD_ANNOTATION_CHAR)
						if len(stok) > 1 {
							k = stok[0]
							ann = stok[1]
						} else {
							k = stok[0]
						}
					}
				*/

				switch strings.ToLower(kk) {
				case TOKEN_PLATFORM:
					if vv != nil {
						if ps, aok := vv.(string); aok {
							errs = parseYamlPlatform(ps, &c, errs)
						} else {
							errs = append(errs, base.NewGinfraErrorA("Malformed PLATFORM.  Expecting a string."))
						}
					} else {
						errs = append(errs, base.NewGinfraErrorA("Empty PLATFORM."))
					}

				case TOKEN_WORKINGDIR:
					var str string
					str, errs = guardString(TOKEN_CMD+":"+TOKEN_WORKINGDIR, vv, errs)
					c.D = append(c.D, WorkingDir{Path: str})

				case TOKEN_REMOVE:
					var r Remove
					r, errs = tokenRemove(vv, errs)
					c.D = append(c.D, r)

				case TOKEN_REPEAT:
					var rp int
					rp, errs = guardInt(TOKEN_CMD+":"+TOKEN_REPEAT, vv, errs)
					c.C.Repeat = Repeat(rp)

				case TOKEN_REPEAT_DELAY:
					c.C.RepeatDelay, errs = guardInt(TOKEN_CMD+":"+TOKEN_REPEAT_DELAY, vv, errs)

				case TOKEN_CMD:
					if vv != nil {
						if cs, aok := vv.(string); aok {
							c.D = append(c.D, Cmd(cs))
							cmdSeen = true
						} else {
							errs = append(errs, base.NewGinfraErrorA("Malformed CMD.  Expecting a string."))
						}
					} else {
						errs = append(errs, base.NewGinfraErrorA("Empty CMD."))
					}

				case TOKEN_FAILOK:
					c.FailOk, errs = guardTruth(TOKEN_CMD+":"+TOKEN_FAILOK, vv, errs)

				case TOKEN_POST:
					c.C.Post, errs = guardTruth(TOKEN_CMD+":"+TOKEN_POST, vv, errs)

				case TOKEN_OUTPUT:
					var o Output
					o, errs = tokenOutput(vv, errs, hasCmd, cmdSeen)
					c.D = append(c.D, o)

				case TOKEN_PING:
					var p Ping
					p, errs = tokenPing(vv, errs, hasCmd, cmdSeen)
					c.D = append(c.D, p)

				case TOKEN_SUCCESS:
					var str string
					str, errs = guardString(TOKEN_CMD+":"+TOKEN_SUCCESS, vv, errs)
					c.C.Success = &str

				case TOKEN_COMMENT:
					break // NOP

				default:
					errs = append(errs, base.NewGinfraErrorA("Unknown directive in COMMAND.", base.LM_TOKEN, kk))
				}
			}

		} else {
			errs = append(errs, base.NewGinfraErrorA("Poorly formed yaml CMD."))
		}
	}
	if len(errs) < 1 {
		var wi WorkflowItem
		wi.Command = &c
		s.Workflow = append(s.Workflow, &wi)
	}
	return errs
}

func parseYamlGinfra(s *Script, in []interface{}, errs []error) []error {
	var (
		c Ginfra
	)

	for _, v := range in {

		if ivv, ok := v.(map[string]interface{}); ok {

			for kk, vv := range ivv {
				switch strings.ToLower(kk) {
				case TOKEN_WORKINGDIR:
					var str string
					str, errs = guardString(TOKEN_GINFRA+":"+TOKEN_WORKINGDIR, vv, errs)
					c.D = append(c.D, WorkingDir{Path: str})

				case TOKEN_REMOVE:
					var r Remove
					r, errs = tokenRemove(vv, errs)
					c.D = append(c.D, r)

				case TOKEN_CMD:
					var str string
					str, errs = guardString(TOKEN_GINFRA+":"+TOKEN_CMD, vv, errs)
					c.D = append(c.D, Cmd(str))

				case TOKEN_POST:
					c.C.Post, errs = guardTruth(TOKEN_GINFRA+":"+TOKEN_POST, vv, errs)

				case TOKEN_DETACH:
					c.Detach, errs = guardTruth(TOKEN_GINFRA+":"+TOKEN_DETACH, vv, errs)

				case TOKEN_OUTPUT:
					var o Output
					o, errs = tokenOutput(vv, errs, true, true)
					c.D = append(c.D, o)

				case TOKEN_PING:
					var p Ping
					p, errs = tokenPing(vv, errs, true, true)
					c.D = append(c.D, p)

				case TOKEN_REPEAT:
					var rp int
					rp, errs = guardInt(TOKEN_GINFRA+":"+TOKEN_REPEAT, vv, errs)
					c.C.Repeat = Repeat(rp)

				case TOKEN_REPEAT_DELAY:
					c.C.RepeatDelay, errs = guardInt(TOKEN_GINFRA+":"+TOKEN_REPEAT_DELAY, vv, errs)

				case TOKEN_SUCCESS:
					var str string
					str, errs = guardString(TOKEN_GINFRA+":"+TOKEN_SUCCESS, vv, errs)
					c.C.Success = &str

				case TOKEN_COMMENT:
					break // NOP

				default:
					errs = append(errs, base.NewGinfraErrorA("Unknown GINFRA directive.", base.LM_TOKEN, kk))
				}
			}
		} else {
			errs = append(errs, base.NewGinfraErrorA("Poorly formed yaml GINFRA."))
		}
	}

	if len(errs) < 1 {
		var wi WorkflowItem
		wi.Ginfra = &c
		s.Workflow = append(s.Workflow, &wi)
	}
	return errs
}

func parseYamlWorkflow(s *Script, in any, errs []error) []error {
	if innn, ok := in.([]interface{}); ok {
		for _, i := range innn {

			if im, ok := i.(map[string]interface{}); ok {
				for k, v := range im {

					var seenCmd bool
					var t []interface{}

					if v != nil && strings.ToLower(k) != TOKEN_COMMENT && strings.ToLower(k) != TOKEN_LOG {
						switch ct := v.(type) {
						case []interface{}:
							t = ct

						case map[string]interface{}:
							t = append(t, map[string]interface{}{TOKEN_COMMENT: "a"})
							for ik, iv := range ct {

								// Make sure the CMD is first in list.
								if !seenCmd && strings.ToLower(ik) == TOKEN_CMD {
									t[0] = map[string]interface{}{ik: iv}
									seenCmd = true
								} else {
									t = append(t, map[string]interface{}{ik: iv})
								}
							}

						default:
							panic("Unknown type for " + k)
						}
					}

					switch strings.ToLower(k) {
					case TOKEN_CMD:
						if v == nil {
							errs = append(errs, base.NewGinfraErrorA("Poorly formed yaml for CMD.  No directives recognized."))
						} else {
							errs = parseYamlCmd(s, t, seenCmd, errs)
						}

					case TOKEN_GINFRA:
						if v == nil {
							errs = append(errs, base.NewGinfraErrorA("Poorly formed yaml for GINFRA.  No directives recognized."))
						} else {
							errs = parseYamlGinfra(s, t, errs)
						}

					case TOKEN_LOG:
						if log, ok := v.(string); ok {
							var wi WorkflowItem
							wi.Log = &log
							s.Workflow = append(s.Workflow, &wi)
						} else {
							errs = append(errs, base.NewGinfraErrorA("LOG must be a string object."))
						}

					case TOKEN_COMMENT:
						break // NOP

					default:
						errs = append(errs, base.NewGinfraErrorA("Unknown WORKFLOW step type.", base.LM_TOKEN, k))
					}
				}

			} else {
				// Bad yaml
				panic(base.NewGinfraError("Bad yaml format.  Expecting a directive in the WORKFLOW list."))
			}
		}
	} else {
		// Bad yaml
		panic(base.NewGinfraError("Bad yaml format.  WORKFLOW should be a list (yaml array)."))
	}
	return errs
}

func parseYaml(in map[string]interface{}) (*Script, []error) {
	var (
		s    Script
		errs []error
	)

	for k, v := range in {
		switch strings.ToLower(k) {
		case TOKEN_META:
			errs = parseYamlMeta(&s, v, errs)

		case TOKEN_WORKFLOW:
			errs = parseYamlWorkflow(&s, v, errs)

		case TOKEN_COMMENT:
			break // NOP

		default:
			errs = append(errs, base.NewGinfraErrorA("Bad top-level section", base.LM_TOKEN, k))
		}

	}

	return &s, errs
}

// LoadAndParseYamlData loads and parses YAML data from the given data and returns a Script object or an error.
// rb is the YAML data as a byte slice, and path is the source file's path for error context and tracking.
// On success, it returns a populated *Script and nil error, otherwise a nil Script and the encountered error.
func LoadAndParseYamlData(rb []byte, path string) (scr *Script, err error) {

	//var s Script
	var (
		s    map[string]interface{}
		errs []error
	)

	if err = yaml.Unmarshal(rb, &s); err != nil {
		return nil, base.NewGinfraErrorChildA("Could not load script yaml.  Bad format.", err,
			base.LM_FILEPATH, path)
	}

	// Bad yaml will stop the parsing.
	defer func() {
		if r := recover(); r != nil {
			if perr, ok := r.(*base.GinfraError); ok {
				err = perr
			} else {
				err = base.NewGinfraErrorA("Parsing failed due to panic.", base.LM_CAUSE, fmt.Sprintf("%v", r))
			}
		}
	}()

	scr, errs = parseYaml(s)
	scr.SourceFile = path
	if len(errs) > 0 {
		err = base.NewGinfraErrorA("Errors parsing yaml.", base.LM_FILEPATH, path, errs)
	}

	return scr, err
}

// LoadAndParseYaml reads a YAML file from the specified path, parses its content, and returns a Script object or an error.
func LoadAndParseYaml(path string) (scr *Script, err error) {

	//var s Script
	var (
		rb []byte
	)

	if rb, err = os.ReadFile(path); err != nil {
		cwd, _ := os.Getwd()
		return nil, base.NewGinfraErrorChildA("Could not read script yaml.", err,
			base.LM_FILEPATH, path, base.LM_WORKINGDIR, cwd)
	}

	return LoadAndParseYamlData(rb, path)
}
