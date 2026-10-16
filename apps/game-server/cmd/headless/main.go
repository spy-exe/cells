// headless exposes bounded, simultaneous simulation steps through local stdio.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/spy-exe/cells/apps/game-server/internal/training"
	"io"
	"os"
)

type request struct {
	Op        string                     `json:"op"`
	Seed      uint64                     `json:"seed"`
	Agents    *int                       `json:"agents,omitempty"`
	Opponents *int                       `json:"opponents,omitempty"`
	Horizon   *int                       `json:"horizon,omitempty"`
	Actions   map[string]training.Action `json:"actions,omitempty"`
	Policy    string                     `json:"policy,omitempty"`
}
type response struct {
	Protocol    int              `json:"protocol"`
	Result      *training.Result `json:"result,omitempty"`
	PolicyNanos int64            `json:"policy_ns,omitempty"`
	Error       string           `json:"error,omitempty"`
	Closed      bool             `json:"closed,omitempty"`
}

func main() { serve(os.Stdin, os.Stdout) }
func serve(input io.Reader, output io.Writer) {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), 65536)
	writer := bufio.NewWriter(output)
	encoder := json.NewEncoder(writer)
	var env *training.Env
	for scanner.Scan() {
		var req request
		var result training.Result
		reply := response{Protocol: 1}
		decoder := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&req)
		if err == nil {
			var extra any
			if decoder.Decode(&extra) != io.EOF {
				err = errors.New("one JSON object required")
			}
		}
		if err == nil {
			switch req.Op {
			case "reset":
				cfg := training.DefaultConfig()
				if req.Agents != nil {
					cfg.Agents = *req.Agents
				}
				if req.Opponents != nil {
					cfg.Opponents = *req.Opponents
				}
				if req.Horizon != nil {
					cfg.Horizon = *req.Horizon
				}
				var candidate *training.Env
				candidate, err = training.New(req.Seed, cfg)
				if err == nil {
					env = candidate
					result = env.Snapshot()
					reply.Result = &result
				}
			case "step":
				if env == nil {
					err = errors.New("reset required")
				} else {
					if req.Policy != "" {
						if len(req.Actions) > 0 {
							err = errors.New("policy and actions are exclusive")
						} else {
							result, err = env.StepPolicy(req.Policy)
						}
					} else {
						result, err = env.Step(req.Actions)
					}
					if err == nil {
						reply.Result = &result
						reply.PolicyNanos = env.LastPolicyNanos
					}
				}
			case "close":
				reply.Closed = true
			default:
				err = errors.New("unknown operation")
			}
		}
		if err != nil {
			reply.Error = err.Error()
		}
		if encoder.Encode(reply) != nil || writer.Flush() != nil {
			return
		}
		if reply.Closed {
			return
		}
	}
}
