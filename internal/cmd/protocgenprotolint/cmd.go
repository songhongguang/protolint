package protocgenprotolint

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/yoheimuta/protolint/internal/cmd/subcmds"

	"github.com/yoheimuta/protolint/internal/cmd/subcmds/lint"

	"github.com/golang/protobuf/proto"
	protogen "github.com/golang/protobuf/protoc-gen-go/plugin"

	"github.com/yoheimuta/protolint/internal/osutil"
)

// Do runs the command logic.
func Do(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) osutil.ExitCode {
	subCmd, err := newSubCmd(stdin, stdout, stderr)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return osutil.ExitFailure
	}
	return subCmd.Run()
}

func newSubCmd(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) (*lint.CmdLint, error) {
	data, err := ioutil.ReadAll(stdin)
	if err != nil {
		return nil, err
	}

	var req protogen.CodeGeneratorRequest
	err = proto.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}

	flags, err := newFlags(req)
	if err != nil {
		return nil, err
	}

	subCmd, err := lint.NewCmdLint(
		*flags,
		stdout,
		stderr,
	)
	if err != nil {
		return nil, err
	}
	return subCmd, nil
}

func newFlags(
	req protogen.CodeGeneratorRequest,
) (*lint.Flags, error) {
	flags, err := lint.NewFlags(req.FileToGenerate)
	if err != nil {
		return nil, err
	}

	var pf subcmds.PluginFlag
	for _, p := range strings.Split(req.GetParameter(), ",") {
		params := strings.SplitN(strings.TrimSpace(p), "=", 2)
		switch params[0] {
		case "":
			continue
		case "config_dir_path":
			if len(params) != 2 {
				return nil, fmt.Errorf("config_dir_path should be specified")
			}
			flags.ConfigDirPath = params[1]
		case "fix":
			flags.FixMode = true
		case "reporter":
			if len(params) != 2 {
				return nil, fmt.Errorf("reporter should be specified")
			}
			value := params[1]
			r, err := lint.GetReporter(value)
			if err != nil {
				return nil, err
			}
			flags.Reporter = r
		case "plugin":
			if len(params) != 2 {
				return nil, fmt.Errorf("plugin should be specified")
			}
			err = pf.Set(params[1])
			if err != nil {
				return nil, err
			}
		case "v":
			flags.Verbose = true
		default:
			return nil, fmt.Errorf("unmatched parameter: %s", p)
		}
	}

	plugins, err := pf.BuildPlugins(flags.Verbose)
	if err != nil {
		return nil, err
	}
	flags.Plugins = plugins

	return &flags, nil
}
