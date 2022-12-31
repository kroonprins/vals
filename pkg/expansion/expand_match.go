package expansion

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type ExpandRegexMatch struct {
	Target *regexp.Regexp
	Lookup func(string) (interface{}, error)
	Only   []string
}

var DefaultRefRegexp = regexp.MustCompile(`((secret)?(ref|fileref))\+([^\+:]*://[^\+]+)\+?`)

func (e *ExpandRegexMatch) InString(s string) (interface{}, error) {
	var sb strings.Builder
	res := make(map[string]interface{})
	for {
		ixs := e.Target.FindStringSubmatchIndex(s)
		if ixs == nil {
			sb.WriteString(s)
			break
		}
		kind := s[ixs[2]:ixs[3]]
		if len(e.Only) > 0 {
			var shouldExpand bool
			for _, k := range e.Only {
				if k == kind {
					shouldExpand = true
					break
				}
			}
			if !shouldExpand {
				sb.WriteString(s)
				break
			}
		}
		ref := s[ixs[8]:ixs[9]]
		val, err := e.Lookup(ref)
		if err != nil {
			return nil, fmt.Errorf("expand %s: %v", ref, err)
		}

		if s[ixs[6]:ixs[7]] == "fileref" {
			fileRefTarget, err := getFileRefTarget(ref)
			if err != nil {
				return nil, err
			}
			sb.WriteString(fileRefTarget)
			err = writeFile(fileRefTarget, val)
			if err != nil {
				return nil, fmt.Errorf("error writing file %s: %v", fileRefTarget, err)
			}
		} else {
			switch typed_val := val.(type) {
			case string:
				sb.WriteString(s[:ixs[0]])
				sb.WriteString(typed_val)
			case map[string]interface{}:
				for k, v := range typed_val {
					res[k] = v
				}
			case map[interface{}]interface{}:
				for k, v := range typed_val {
					res[fmt.Sprintf("%v", k)] = v
				}
			default:
				return nil, fmt.Errorf("unexpected output format for %s: %#v", ref, val)
			}
		}

		s = s[ixs[1]:]
	}
	if len(res) > 0 {
		if sb.Len() > 0 {
			return nil, fmt.Errorf("when combining references with '+' all must evaluate to the same type")
		}
		return res, nil
	}
	return sb.String(), nil

}

func (e *ExpandRegexMatch) InMap(target map[string]interface{}) (map[string]interface{}, error) {
	ret, err := ModifyStringValues(target, func(p string) (interface{}, error) {
		ret, err := e.InString(p)
		if err != nil {
			return nil, err
		}
		return ret, nil
	})

	if err != nil {
		return nil, err
	}

	switch typed_ret := ret.(type) {
	case map[string]interface{}:
		return typed_ret, nil
	default:
		return nil, fmt.Errorf("unexpected type: %v: %T", ret, ret)
	}
}

func getFileRefTarget(s string) (string, error) {
	uri, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	queryParams := uri.Query()
	if !queryParams.Has("fileref_target") {
		return "", fmt.Errorf("fileref requires a query parameter 'fileref_target': %s", s)
	}
	return queryParams.Get("fileref_target"), nil
}

func writeFile(fileTargetRef string, val interface{}) error {
	var data []byte
	switch typed_val := val.(type) {
	case string:
		data = []byte(typed_val)
	case map[string]interface{}:
		bytes, err := yaml.Marshal(typed_val)
		if err != nil {
			return err
		}
		data = bytes
	default:
		return fmt.Errorf("unexpected output format: %#v", val)
	}

	return os.WriteFile(fileTargetRef, data, 0666)
}
