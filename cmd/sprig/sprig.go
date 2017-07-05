package sprig

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/cmd/helm/strvals"

	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"
)

var Version = "unknown"

type sprigCommand struct {
	valueFiles valueFiles
	envValues  bool
	dryRun     bool
	version    bool
	values     []string
	target     string
}

func NewSprigCmd() *cobra.Command {
	sprigCmd := &sprigCommand{}

	sprigCLI := &cobra.Command{
		Use:   "sprig",
		Short: "A CLI for golang text/template processing",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				sprigCmd.target = args[0]
			} else if len(args) > 1 {
				return fmt.Errorf("must provide only one file to template")
			}
			return sprigCmd.run()
		},
	}
	sprigCLI.Flags().BoolVar(&sprigCmd.envValues, "env", false, "pull template values from the environment")
	sprigCLI.Flags().BoolVar(&sprigCmd.version, "version", false, "print version and exit")
	sprigCLI.Flags().VarP(&sprigCmd.valueFiles, "values", "f", "specify values in YAML file (can specify multiple, comma separated)")
	sprigCLI.Flags().StringArrayVar(&sprigCmd.values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")

	return sprigCLI
}

type valueFiles []string

func (v *valueFiles) String() string {
	return fmt.Sprint(*v)
}

func (v *valueFiles) Type() string {
	return "valueFiles"
}

func (v *valueFiles) Set(value string) error {
	for _, filePath := range strings.Split(value, ",") {
		*v = append(*v, filePath)
	}
	return nil
}

func (i *sprigCommand) vals() (map[string]interface{}, error) {
	base := map[string]interface{}{}

	// User specified a values files via -f/--values
	for _, filePath := range i.valueFiles {
		currentMap := map[string]interface{}{}
		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	// User specified a value via --set
	for _, value := range i.values {
		if err := strvals.ParseInto(value, base); err != nil {
			return nil, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	// Environment set stuff
	if i.envValues {
		envMap := map[string]interface{}{}
		envVars := os.Environ()
		for _, envVar := range envVars {
			splitVar := strings.SplitN(envVar, "=", 2)
			envMap[splitVar[0]] = splitVar[1]
		}
		base = mergeValues(base, envMap)
	}

	return base, nil
}

func (i *sprigCommand) run() error {
	if i.version {
		fmt.Println("sprig: " + Version)
		return nil
	}

	vals, err := i.vals()
	if err != nil {
		return err
	}

	var r io.Reader
	if shouldReadStdin() {
		r = os.Stdin
	} else {
		if i.target == "" {
			return fmt.Errorf("must provide a file to template")
		}
		r, err = os.Open(i.target)
		if err != nil {
			return err
		}
	}

	templateData, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("could not read input: %v", err)
	}

	tmpl, err := template.New("gotmpl").Funcs(sprig.TxtFuncMap()).Parse(string(templateData))
	if err != nil {
		return fmt.Errorf("could not parse template: %v", err)
	}
	tmpl.Option("missingkey=error")

	return tmpl.Execute(os.Stdout, vals)
}

// Merges source and destination map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = nextMap
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

// shouldReadStdin determines if stdin should be considered a valid source of data for templating.
func shouldReadStdin() bool {
	stdinStat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	return stdinStat.Mode()&os.ModeCharDevice == 0
}
