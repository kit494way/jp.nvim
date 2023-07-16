package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmespath-community/go-jmespath"
	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

var resultBufs *resultBuffers = &resultBuffers{buffers: make(map[nvim.Buffer]nvim.Buffer)}

type resultBuffers struct {
	buffers map[nvim.Buffer]nvim.Buffer
}

func (r *resultBuffers) buffer(v *nvim.Nvim, k nvim.Buffer) (buf nvim.Buffer, err error) {
	if err = r.clean(v); err != nil {
		return buf, err
	}

	buf, ok := r.buffers[k]
	if !ok {
		// create scratch buffer
		buf, err = v.CreateBuffer(false, true)
		if err != nil {
			return buf, err
		}

		if err = v.SetBufferOption(buf, "filetype", "json"); err != nil {
			return buf, err
		}

		r.buffers[k] = buf
	}

	return buf, nil
}

func (r *resultBuffers) deleteBuffer(v *nvim.Nvim, k nvim.Buffer) error {
	if err := v.DeleteBuffer(r.buffers[k], map[string]bool{"force": true, "unload": false}); err != nil {
		return err
	}
	delete(r.buffers, k)
	return nil
}

func (r *resultBuffers) clean(v *nvim.Nvim) error {
	for k, b := range r.buffers {
		valid, err := v.IsBufferValid(b)
		if err != nil {
			return err
		}

		// the buffer is wiped out
		if !valid {
			delete(r.buffers, k)
			continue
		}

		loaded, err := v.IsBufferLoaded(b)
		if err != nil {
			return err
		}

		// the buffer may be removed from the buffer list
		if !loaded {
			if err := r.deleteBuffer(v, k); err != nil {
				return err
			}
			continue
		}

		// If the buffer is hidden and the key of that is invalid, delete the item.
		hidden, err := isBufferHidden(v, b)
		if err != nil {
			return err
		}
		if hidden {
			valid, err := v.IsBufferValid(k)
			if err != nil {
				return err
			}
			if !valid {
				if err := r.deleteBuffer(v, k); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type encoder struct {
	*json.Encoder
	buf *bytes.Buffer
}

func newEncoder() *encoder {
	buf := &bytes.Buffer{}
	e := json.NewEncoder(buf)
	e.SetEscapeHTML(false)
	e.SetIndent("", "  ")

	return &encoder{
		Encoder: e,
		buf:     buf,
	}
}

func jsonUmarshalBuffer(v *nvim.Nvim, b nvim.Buffer, data any) error {
	lines, err := v.BufferLines(b, 0, -1, false)
	if err != nil {
		return err
	}

	input := bytes.Join(lines, []byte("\n"))
	return json.Unmarshal(input, data)
}

func jsonMarshal(data any) ([]byte, error) {
	e := newEncoder()
	if err := e.Encode(data); err != nil {
		return nil, err
	}

	return bytes.TrimSuffix(e.buf.Bytes(), []byte("\n")), nil
}

func isBufferHidden(v *nvim.Nvim, b nvim.Buffer) (bool, error) {
	res, err := v.Exec(fmt.Sprintf("echo empty(win_findbuf(%d))", b), map[string]interface{}{"output": true})
	if err != nil {
		return false, err
	}

	output, isStr := res["output"].(string)
	if !isStr {
		return false, errors.New("Unexpected output")
	}

	return output == "1", nil
}

func displayResult(v *nvim.Nvim, curBuf nvim.Buffer, data any) error {
	bs, err := jsonMarshal(data)
	if err != nil {
		return err
	}

	buf, err := resultBufs.buffer(v, curBuf)
	if err != nil {
		return err
	}

	if hidden, err := isBufferHidden(v, buf); err == nil && hidden {
		curWin, err := v.CurrentWindow()
		if err != nil {
			return err
		}

		// open the result buffer
		if _, err = v.Exec(fmt.Sprintf("bo vs | b %d", buf), map[string]interface{}{"output": false}); err != nil {
			return err
		}

		if err = v.SetCurrentWindow(curWin); err != nil {
			return err
		}
	}

	return v.SetBufferLines(buf, 0, -1, true, bytes.Split(bs, []byte("\n")))
}

func search(v *nvim.Nvim, args []string) error {
	query := args[0]

	curBuf, err := v.CurrentBuffer()
	if err != nil {
		return err
	}

	var data any
	if err := jsonUmarshalBuffer(v, curBuf, &data); err != nil {
		return err
	}

	res, err := jmespath.Search(query, data)
	if err != nil {
		return err
	}

	return displayResult(v, curBuf, res)
}

func main() {
	plugin.Main(func(p *plugin.Plugin) error {
		p.HandleCommand(&plugin.CommandOptions{Name: "JP", NArgs: "1"}, search)

		return nil
	})
}
