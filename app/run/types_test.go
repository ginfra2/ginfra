/*
Package run
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Test for types.
*/
package run

import (
	"github.com/stretchr/testify/assert"
	"os/exec"
	"reflect"
	"testing"
)

var (
	ok      = "sdfsdfsfs"
	nothing = map[string]interface{}{}
	broke   = "sdfsdfsdf <<<sdfsdfsf>sdf"
)

func Test_TypesGetPlatform(t *testing.T) {

	p, err := getPlatform("windows")
	assert.NoError(t, err)
	assert.Equal(t, PWindows, p, "windows")

	p, err = getPlatform("blarg")
	assert.Error(t, err)
}

func TestCommandPrepare(t *testing.T) {
	// Define the test cases
	tests := []struct {
		name    string
		cmd     Command
		v       map[string]interface{}
		wantErr bool
	}{
		{
			name:    "Valid Prepare",
			cmd:     Command{D: []interface{}{Cmd("sdfasdf <<<key>>> asdfasdf")}},
			v:       map[string]interface{}{"key": "value"},
			wantErr: false,
		},
		{
			name:    "Invalid Prepare",
			cmd:     Command{D: []interface{}{Cmd("sdfasdf <<<key asdfasdf")}},
			v:       map[string]interface{}{"key": "value"},
			wantErr: true,
		},
		// Add more test cases as needed
	}

	sctx := ScriptContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calling the function to test it
			sctx.Configs = tt.v
			if err := tt.cmd.Prepare(&sctx); (err != nil) != tt.wantErr {
				t.Errorf("Command.Prepare() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandPrepareCases(t *testing.T) {

	cc := CommonDirectives{}
	sctx := ScriptContext{}

	type args struct {
		d []interface{}
		c *CommonDirectives
		v map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    []interface{}
		wantErr bool
	}{
		{
			name: "Cmd Replaces",
			args: args{
				d: []interface{}{Cmd("hello")},
				c: &cc,
				v: map[string]interface{}{},
			},
			want:    []interface{}{Cmd("hello")},
			wantErr: false,
		},
		{
			name: "WorkingDir",
			args: args{
				d: []interface{}{WorkingDir{Path: "aaa <<<AAA>>> aaa"}},
				c: &cc,
				v: map[string]interface{}{"AAA": "BBB"},
			},
			want:    []interface{}{WorkingDir{Path: "aaa BBB aaa"}},
			wantErr: false,
		},
		{
			name: "WorkingDir Error",
			args: args{
				d: []interface{}{WorkingDir{Path: "aaa <<<AAA aaa"}},
				c: &cc,
				v: map[string]interface{}{"AAA": "BBB"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Remove",
			args: args{
				d: []interface{}{Remove{Paths: []string{"aaa <<<AAA>>> aaa", "bbb"}}},
				c: &cc,
				v: map[string]interface{}{"AAA": "BBB"},
			},
			want:    []interface{}{Remove{Paths: []string{"aaa BBB aaa", "bbb"}}},
			wantErr: false,
		},
		{
			name: "Remove Error",
			args: args{
				d: []interface{}{Remove{Paths: []string{"aaa <<<AAA> aaa", "bbb"}}},
				c: &cc,
				v: map[string]interface{}{"AAA": "BBB"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Output",
			args: args{
				d: []interface{}{Output{D: []interface{}{Delay(1), Until(200), Source("aaa <<<AAA>>> aaa")}}},
				c: &cc,
				v: map[string]interface{}{"AAA": "BBB"},
			},
			want:    []interface{}{Output{D: []interface{}{Delay(1), Until(200), Source("aaa BBB aaa")}}},
			wantErr: false,
		},
		{
			name: "Output Error 1",
			args: args{
				d: []interface{}{Output{D: []interface{}{Delay(1), Source("aaa <<<AAA aaa")}}},
				c: &cc,
				v: map[string]interface{}{"AAA": "BBB"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Output Error Scan",
			args: args{
				d: []interface{}{Output{D: []interface{}{Scan("aaa <<<AAA aaa")}}},
				c: &cc,
				v: map[string]interface{}{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Output Error NoScan",
			args: args{
				d: []interface{}{Output{D: []interface{}{NoScan("aaa <<<AAA aaa")}}},
				c: &cc,
				v: map[string]interface{}{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Output Error Extract Target",
			args: args{
				d: []interface{}{Output{D: []interface{}{Extract{Target: "sdf <<<sdf sdfsdf", Regex: "sadfdsf"}}}},
				c: &cc,
				v: map[string]interface{}{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Output Error Capture Target",
			args: args{
				d: []interface{}{Output{D: []interface{}{Capture{Target: "sdf <<<sdf sdfsdf"}}}},
				c: &cc,
				v: map[string]interface{}{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Output Error Extract Reg",
			args: args{
				d: []interface{}{Output{D: []interface{}{Extract{Target: "sdsdf sdfsdf", Regex: "sadf <<< dsf"}}}},
				c: &cc,
				v: map[string]interface{}{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Ping",
			args: args{
				d: []interface{}{Ping{Url: "http://<<<HOST>>>:8080/"}},
				c: &cc,
				v: map[string]interface{}{"HOST": "localhost"},
			},
			want:    []interface{}{Ping{Url: "http://localhost:8080/"}},
			wantErr: false,
		},
		{
			name: "Ping Error",
			args: args{
				d: []interface{}{Ping{Url: "http://<<<HOST:8080/"}},
				c: &cc,
				v: map[string]interface{}{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Output Capture Target",
			args: args{
				d: []interface{}{Output{D: []interface{}{Capture{Target: "sdf"}}}},
				c: &cc,
				v: map[string]interface{}{},
			},
			want:    []interface{}{Output{D: []interface{}{Capture{Target: "sdf"}}}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sctx.Configs = tt.args.v
			got, err := Prepare(&sctx, tt.args.d, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Prepare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Prepare() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("Success error", func(t *testing.T) {
		var s = "hi <<< sdf"
		cc.Success = &s
		_, err := Prepare(&sctx, []interface{}{}, &cc)
		assert.Error(t, err)
	})

	t.Run("Directive panic", func(t *testing.T) {
		assert.Panics(t, func() { _, _ = Prepare(&sctx, []interface{}{1}, &cc) })
	})

	t.Run("Command", func(t *testing.T) {
		var (
			c   Command
			cmd exec.Cmd
		)
		c.SetCommand(&cmd)
		assert.Equal(t, &cmd, c.GetCommand())
		c.SetKilled(true)
		assert.Equal(t, true, c.IsKilled())
	})

	t.Run("GinfraInstance", func(t *testing.T) {
		var (
			c   GinfraInstance
			cmd exec.Cmd
		)
		c.SetCommand(&cmd)
		assert.Equal(t, &cmd, c.GetCommand())
		c.SetKilled(true)
		assert.Equal(t, true, c.IsKilled())
	})

	t.Run("Ginfra", func(t *testing.T) {
		var (
			c Ginfra
			s = "aaabbbccc <<< ssss"
		)
		c.C.Success = &s
		assert.Error(t, c.Prepare(&sctx))
	})

}
