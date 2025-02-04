package vals

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func Inputs(f string) ([]yaml.Node, error) {
	var reader io.Reader
	if f == "-" {
		reader = os.Stdin
	} else if f != "" {
		fp, err := os.Open(f)
		if err != nil {
			return nil, err
		}
		reader = fp
		defer fp.Close()
	} else {
		return nil, fmt.Errorf("nothing to eval: No file specified")
	}
	return nodesFromReader(reader)
}

func nodesFromReader(reader io.Reader) ([]yaml.Node, error) {
	nodes := []yaml.Node{}
	buf := bufio.NewReader(reader)
	decoder := yaml.NewDecoder(buf)
	for {
		node := yaml.Node{}
		if err := decoder.Decode(&node); err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func Output(output io.Writer, format string, nodes []yaml.Node) error {
	for i, node := range nodes {
		var v interface{}
		if err := node.Decode(&v); err != nil {
			return err
		}
		if format == "json" {
			bs, err := json.Marshal(v)
			if err != nil {
				return err
			}
			fmt.Fprintln(output, string(bs))
		} else {
			encoder := yaml.NewEncoder(output)
			encoder.SetIndent(2)

			if err := encoder.Encode(v); err != nil {
				return err
			}
		}
		if i != len(nodes)-1 {
			fmt.Fprintln(output, "---")
		}
	}
	return nil
}
