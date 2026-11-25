/*
Package common
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Various tools and helpers.  They should all be obvious in what they do.
*/
package common

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gitlab.com/ginfra/ginfra/base"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func GetFileAsString(filepath string) (string, error) {
	d, err := os.ReadFile(filepath)
	return string(d), err
}

func CopyFile(src, dest string) error {

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = s.Close() }()
	st, _ := os.Stat(src)

	d, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = d.Close() }()

	_, err = io.Copy(d, s)
	_ = d.Close()

	if err == nil {
		err = os.Chmod(src, st.Mode())
	}

	return err
}

func GetInterfaceName(serviceName string) string {
	return fmt.Sprintf("service_%v.yaml", serviceName)
}

const FMARKER_DONT_UPDATE = "!"
const FMARKER_DONT_TEMPLATE = "^"
const FMARKER_DONT_UPDATE_OR_TEMPLATE = "#"

const FileFlag_Hide = 0b1
const FileFlag_NoUpdate = 0b10
const FileFlag_NoTemplate = 0b100

func GetDestFilepath(destRoot, srcRoot, srcPath string, handleHideFiles, handleNoUpdateFiles,
	handleNoTemplateFiles bool) (string, int) {
	var flags int

	df := filepath.Join(destRoot, srcPath[len(srcRoot):])
	b := filepath.Base(srcPath)

	if handleHideFiles {
		if strings.HasPrefix(b, "_") {
			df = strings.ReplaceAll(df, b, b[1:])
			flags = flags + FileFlag_Hide
		}
		if strings.HasSuffix(b, ".got") {
			df = df[:len(df)-4] + ".go"
			flags = flags + FileFlag_Hide
		}
	}

	if handleNoUpdateFiles {
		if strings.HasPrefix(b, FMARKER_DONT_UPDATE) || strings.HasPrefix(b, FMARKER_DONT_UPDATE_OR_TEMPLATE) {
			df = strings.ReplaceAll(df, b, b[1:])
			flags = flags + FileFlag_NoUpdate
		}
	}
	if handleNoTemplateFiles {
		if strings.HasPrefix(b, FMARKER_DONT_TEMPLATE) || strings.HasPrefix(b, FMARKER_DONT_UPDATE_OR_TEMPLATE) {
			df = strings.ReplaceAll(df, b, b[1:])
			flags = flags + FileFlag_NoTemplate
		}
	}
	return df, flags
}

func MakeDir(path string) error {
	var (
		info os.FileInfo
		err  error
	)

	if err = os.MkdirAll(path, 0755); err != nil {
		if info, err = os.Stat(path); err == nil {
			if info != nil {
				if !info.IsDir() {
					err = base.NewGinfraErrorChildA("Could not create directory.  Path already exists as a file.",
						err, base.LM_FILEPATH, path)
				}
			} else {
				panic("If error is nil the info should be populated.")
			}

		} else {
			err = base.NewGinfraErrorChildA("Could not create directory.", err, base.LM_FILEPATH, path)
		}
	}

	return err
}

func CopyFiles(srcRoot, destRoot string, handleHideFiles, handleNoUpdateFiles, handleNoTemplateFiles bool) error {

	if err := MakeDir(destRoot); err != nil {
		return base.NewGinfraErrorChildA("Could not create destination directory.", err, base.LM_PATH, destRoot)
	}

	err := filepath.Walk(srcRoot,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				if path != srcRoot {
					dir := filepath.Join(destRoot, path[len(srcRoot):])
					err = MakeDir(dir)
				}

			} else {
				df, _ := GetDestFilepath(destRoot, srcRoot, path, handleHideFiles, handleNoUpdateFiles,
					handleNoTemplateFiles)
				err = CopyFile(path, df)
			}

			return err
		})

	return err
}

// Base64ToJsonMap decodes a Base64-encoded string and parses it into a JSON map.
// Returns the parsed map and an error if decoding or parsing fails.
func Base64ToJsonMap(data string) (map[string]interface{}, error) {
	var r map[string]interface{}
	d, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, base.NewGinfraErrorChild("Bad base64.", err)
	}
	err = json.Unmarshal(d, &r)
	if err != nil {
		return nil, base.NewGinfraErrorChild("Bad json.", err)
	}
	return r, nil
}

// JsonMapToBase64 converts a map with string keys and interface{} values into a Base64-encoded string representation.
func JsonMapToBase64(data map[string]interface{}) string {
	j, err := json.Marshal(data)
	if err != nil {
		panic("Json failed to marshal map.  This should never happen.")
	}
	return base64.StdEncoding.EncodeToString(j)
}

// StringToJsonmap converts a JSON string into a map[string]interface{} representation and returns it along with any error.
func StringToJsonmap(data string) (map[string]interface{}, error) {
	var r map[string]interface{}
	err := json.Unmarshal([]byte(data), &r)
	if err != nil {
		return nil, base.NewGinfraErrorChild("Bad json.", err)
	}
	return r, nil
}

// Path2String concatenates multiple string arguments with a dot (".") separator into a single string.
func Path2String(path ...string) string {
	var sb strings.Builder
	for i, s := range path {
		sb.WriteString(s)
		if i < len(path)-1 {
			sb.WriteString(".")
		}
	}
	return sb.String()
}

// GetActualLocalhostIf will return the actual hostname of the host if the host is either 'localhost' or loopback.
// Using either of those can result in being put on an interface that other nodes can't contact.
func GetActualLocalhostIf(host string) string {
	/*
		var err error
		if strings.ToLower(host) == "localhost" || host == "127.0.0.1" {
			if host, err = os.Hostname(); err != nil {
				panic(fmt.Sprintf("Could not get hostname from OS.  %v", err))
			}
		}*/
	return host
}

// GetOutboundIP determines the preferred outbound IP address of the local machine and returns it as a string.
func GetOutboundIP() string {
	c, err := net.Dial("udp", "1.1.1.1:20")
	if err != nil {
		return ""
	}
	defer func(c net.Conn) {
		err := c.Close()
		if err != nil {

		}
	}(c)
	return c.LocalAddr().(*net.UDPAddr).IP.String()
}

// ReadAllString reads the content of the file at the specified path and returns it as a string.
// It returns an error if the file cannot be read.
func ReadAllString(path string) (string, error) {
	var r string

	d, err := os.ReadFile(path)
	if err == nil {
		r = string(d)
	}

	return r, err
}

// ReadJsonFileMap reads a JSON file from the given path and unmarshals its content into the provided map target.
// It returns an error if the file cannot be read or if the unmarshaling fails.
// The "kind" parameter is used in error messages to indicate the type of file being processed.
func ReadJsonFileMap(path, kind string, target *map[string]interface{}) error {
	d, err := os.ReadFile(path)
	if err != nil {
		return base.NewGinfraErrorChildA("Could not open "+kind+" file", err,
			base.LM_FILEPATH, path)
	}
	err = json.Unmarshal(d, &target)
	if err != nil {
		return base.NewGinfraErrorChildA("Could not get json from  "+kind+" file", err,
			base.LM_FILEPATH, path)
	}
	return err
}

// WriteJsonFileA writes the given struct as a JSON file to the specified path and assigns a type name for error context.
// It returns an error if the file cannot be written or if an issue occurs during JSON serialization.
func WriteJsonFileA(jsonStruct any, path string, kind string) error {
	d, _ := json.Marshal(jsonStruct)
	err := os.WriteFile(path, d, 0644)
	if err != nil {
		err = base.NewGinfraErrorChildA(fmt.Sprintf("Could not write %s.", kind), err, base.LM_FILEPATH, path)
	}
	return err
}

// ReadJsonFile reads a JSON file from the given path and unmarshals its contents into a value of type T.
// Returns the unmarshaled value or an error if the file cannot be read or parsed.
// The `kind` parameter is used for contextual error messages.
func ReadJsonFile[T comparable](path string, kind string) (T, error) {
	var def T
	d, err := os.ReadFile(path)
	if err != nil {
		return def, base.NewGinfraErrorChildA("Could not open "+kind+" file", err,
			base.LM_FILEPATH, path)
	}
	err = json.Unmarshal(d, &def)
	if err != nil {
		return def, base.NewGinfraErrorChildA("Could not get json from  "+kind+" file", err,
			base.LM_FILEPATH, path)
	}
	return def, err
}

// WriteJsonFile writes the given data in JSON format to the specified file path with appropriate permissions.
// If an error occurs during the file writing process, a GinfraError is returned with additional context information.
func WriteJsonFile[T comparable](data T, path string, kind string) error {
	d, _ := json.Marshal(data)
	err := os.WriteFile(path, d, 0644)
	if err != nil {
		err = base.NewGinfraErrorChildA(fmt.Sprintf("Could not write %s.", kind), err, base.LM_FILEPATH,
			path)
	}
	return err
}

const (
	StateReplacerOpen = iota
	StateReplacerEnter1
	StateReplacerEnter2
	StateReplacerExit
	StateReplacerErrorDangle
)

const REPLACER_ENTER = '<'
const REPLACER_EXIT = '>'

func simpleReplacerSnarfer(rr *strings.Reader) (string, error, int) {
	var (
		c     rune
		err   error
		b     strings.Builder
		state = StateReplacerOpen
	)

	for {
		if c, _, err = rr.ReadRune(); err != nil {
			if err == io.EOF {
				err = nil
			}
			break

		} else {

			switch state {

			case StateReplacerOpen:
				if c == REPLACER_EXIT {
					state = StateReplacerEnter1
				} else {
					b.WriteRune(c)
				}

			case StateReplacerEnter1:
				if c == REPLACER_EXIT {
					state = StateReplacerEnter2
				} else {
					b.WriteRune(REPLACER_EXIT)
					b.WriteRune(c)
					state = StateReplacerOpen
				}

			case StateReplacerEnter2:
				if c == REPLACER_EXIT {
					state = StateReplacerExit
				} else {
					b.WriteRune(REPLACER_EXIT)
					b.WriteRune(REPLACER_EXIT)
					b.WriteRune(c)
					state = StateReplacerOpen
				}

			default:
				panic("BUG: this should never happen.")
			}

			if state == StateReplacerExit {
				break
			}
		}
	}

	return strings.TrimSpace(b.String()), err, state
}

// SimpleReplacerHandler takes a string and a map of replacement values.
// It iterates through the string and replaces every instance of a replacement
// key surrounded by double parentheses with its corresponding value from the
// map. If any replacements are left dangling or if there is an error during
// the replacement process, an error is returned. Otherwise, the function
// returns the modified string and no error.  It uses <<< and >>> to mark the open
// and close of the replacement.  Why that?  Because I want to leave mustache replacements intact
// and two (<< >>) would get confused by source code that uses it.
func SimpleReplacerHandler(s string, values func(string) string, retainMiss bool) (string, error) {
	if s == "" {
		return "", nil
	}

	sr := strings.NewReader(s)
	var (
		c    rune
		err  error
		b    strings.Builder
		name string
		mc   int
		num  int

		state = StateReplacerOpen
	)

	for {
		if c, _, err = sr.ReadRune(); err != nil || mc == StateReplacerErrorDangle {
			if err == io.EOF {
				err = nil
			}
			break

		} else {

			switch state {

			case StateReplacerOpen:
				if c == REPLACER_ENTER {
					state = StateReplacerEnter1
				} else {
					b.WriteRune(c)
				}

			case StateReplacerEnter1:
				if c == REPLACER_ENTER {
					state = StateReplacerEnter2
				} else {
					b.WriteRune(REPLACER_ENTER)
					b.WriteRune(c)
					state = StateReplacerOpen
				}

			case StateReplacerEnter2:
				if c == REPLACER_ENTER {
					// snarf
					name, err, mc = simpleReplacerSnarfer(sr)
					if err != nil || mc != StateReplacerExit {
						mc = StateReplacerErrorDangle
						break
					}

					// Emit
					if len(name) > 0 {
						v := values(name)
						if v != "" {
							b.WriteString(fmt.Sprintf("%v", v)) // Super lazy!
							num = num + 1

						} else if retainMiss {
							b.WriteRune(REPLACER_ENTER)
							b.WriteRune(REPLACER_ENTER)
							b.WriteRune(REPLACER_ENTER)
							b.WriteString(name)
							b.WriteRune(REPLACER_EXIT)
							b.WriteRune(REPLACER_EXIT)
							b.WriteRune(REPLACER_EXIT)

						} else {
							// replacing it with empty is altering the data,
							num = num + 1
						}
					}

				} else {
					b.WriteRune(REPLACER_ENTER)
					b.WriteRune(REPLACER_ENTER)
					b.WriteRune(c)
				}
				state = StateReplacerOpen

			default:
				panic("BUG: this should be impossible.")
			}

		} // end if error.
	}

	if err != nil {
		return s, base.NewGinfraErrorChildA("Bad string processing for replacement.", err, base.LM_TEXT, s)
	} else if mc == StateReplacerErrorDangle {
		// Dangling.
		return s, base.NewGinfraErrorA("Dangling replacement '<<< >>>' in string.", base.LM_TEXT, s)
	} else if num == 0 {
		return s, nil
	}

	rs := b.String()
	return rs, nil
}

// SimpleReplacer takes a pointer to a string and a map of replacement values.
// It iterates through the string and replaces every instance of a replacement
// key surrounded by double parentheses with its corresponding value from the
// map. If any replacements are left dangling or if there is an error during
// the replacement process, an error is returned. Otherwise, the function
// returns a pointer to the modified string and no error.
func SimpleReplacer(s string, v map[string]interface{}, retainMiss bool) (string, error) {

	return SimpleReplacerHandler(s, func(in string) string {
		var r string
		if val, ok := v[in]; ok {
			r = fmt.Sprintf("%v", val)
		}
		return r
	}, retainMiss)
}

func CountingSimpleReplacer(s string, v map[string]interface{}, retainMiss bool) (int, string, error) {
	
	count := 0
	s, err := SimpleReplacerHandler(s, func(in string) string {
		count++
		var r string
		if val, ok := v[in]; ok {
			r = fmt.Sprintf("%v", val)
		}
		return r
	}, retainMiss)
	
	return count, s, err
}


// CleanUrl ensures a URL has no redundant "//" after the protocol section and normalizes the path separators.
func CleanUrl(url string) string {
	rParts := strings.SplitN(url, "//", 2)
	if len(rParts) == 2 {
		url = rParts[0] + "//" + strings.ReplaceAll(rParts[1], "//", "/")
	} else {
		url = strings.ReplaceAll(url, "//", "/")
	}
	return url
}

// ReadFileIntoArray reads the contents of a file specified by the path and returns its lines as a slice of strings.
func ReadFileIntoArray(path string) ([]string, error) {
	fr, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(fr)

	var r []string
	scanner := bufio.NewScanner(fr)
	for scanner.Scan() {
		r = append(r, scanner.Text())
	}
	return r, scanner.Err()
}

// ReplaceNewlinesInQuotes scans JSON content and replaces any newlines found between double quotes with \n
func ReplaceNewlinesInQuotes(jc string) string {
	var result strings.Builder
	result.Grow(len(jc))

	inQuotes := false
	escaped := false

	for i := 0; i < len(jc); i++ {
		ch := jc[i]

		if escaped {
			// Previous character was backslash, write current char as-is
			result.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			// Mark that next character is escaped
			result.WriteByte(ch)
			escaped = true
			continue
		}

		if ch == '"' {
			// Toggle quote state
			inQuotes = !inQuotes
			result.WriteByte(ch)
			continue
		}

		if inQuotes && (ch == '\n' || ch == '\r') {
			// Replace newline with \n when inside quotes
			if ch == '\r' && i+1 < len(jc) && jc[i+1] == '\n' {
				// Handle \r\n as a single newline
				result.WriteString("\\n")
				i++ // Skip the \n
			} else if ch == '\n' {
				result.WriteString("\\n")
			}
			// Skip standalone \r
			continue
		}

		result.WriteByte(ch)
	}

	return result.String()
}
