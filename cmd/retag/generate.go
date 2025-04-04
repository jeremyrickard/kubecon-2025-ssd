package retag

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	generateName             = "generate"
	generateShortDescription = "Generate Github Matrix variable for retagging"
	generateLongDescription  = `Generates the json representation of the provided configuration file(s) which will be set as a variable and used in the matrix declaration.
This allows us to spawn a retag job per repository allowing us to scale the retag workflow horizontally`
)

type RetagConfig struct {
	Images []Retag `yaml:"images"`
}

type Retag struct {
	Source      string   `yaml:"source"`
	Destination string   `yaml:"destination"`
	Tags        []string `yaml:"tags"`
}

type generateCmd struct {
	configFile []string
	retags     []Retag
	prefix     string
}

func newGenerateCommand() *cobra.Command {
	gc := &generateCmd{}
	generateCmd := &cobra.Command{
		Use:     generateName,
		Short:   generateShortDescription,
		Long:    generateLongDescription,
		PreRunE: gc.validate,
		RunE:    gc.run,
	}

	f := generateCmd.Flags()
	f.StringSliceVarP(&gc.configFile, "config", "c", []string{"retag.yml"}, "Configuration used to map source repository to desination and the tags to import")
	f.StringVarP(&gc.prefix, "prefix", "p", "mirror", "prefix to use for mapping")
	_ = generateCmd.MarkFlagRequired("config")
	return generateCmd
}

func (gc *generateCmd) validate(_ *cobra.Command, _ []string) error {
	var retags []Retag
	for _, configFile := range gc.configFile {
		data, err := gc.load(configFile)
		if err != nil {
			return err
		}

		rts, err := gc.parse(data)
		if err != nil {
			return err
		}
		retags = append(retags, rts...)
	}
	gc.retags = retags

	return nil
}

func (gc *generateCmd) run(_ *cobra.Command, _ []string) error {
	matrix := gc.generateMatrix()
	data, err := json.Marshal(matrix)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// generateMatrix generates the matrix variable for the retag workflow.
func (gc *generateCmd) generateMatrix() map[string]map[string]string {
	matrix := make(map[string]map[string]string)
	for _, retag := range gc.retags {
		retag := retag
		matrix[sanitizeJobName(&retag)] = map[string]string{
			"source":      retag.Source,
			"destination": retag.Destination,
			"tags":        strings.Join(retag.Tags, ","),
		}
	}

	return matrix
}

// load reads the contents of a file and returns the bytes.
func (gc *generateCmd) load(file string) ([]byte, error) {
	filebytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return filebytes, nil
}

// parse parses the yaml data and returns a list of retags.
func (gc *generateCmd) parse(filebytes []byte) ([]Retag, error) {
	var config RetagConfig
	err := yaml.Unmarshal(filebytes, &config)
	if err != nil {
		return nil, err
	}

	var rts []Retag
	for _, retagConfig := range config.Images {
		if retagConfig.Source == "" {
			return nil, fmt.Errorf("at least one tag is required for retagging %s", retagConfig.Source)
		} else if len(retagConfig.Tags) < 1 {
			return nil, fmt.Errorf("at least one tag is required for retagging %s", retagConfig.Source)
		}

		if retagConfig.Destination == "" {
			retagConfig.Destination = fmt.Sprintf("%s/%s", gc.prefix, retagConfig.Source)
		}
		rts = append(rts, retagConfig)
	}
	return rts, err
}

// sanitizeJobName converts a retag struct to a valid ADO job name.
func sanitizeJobName(retag *Retag) string {
	jobName := retag.Source
	for _, sep := range []string{"/", "-", "."} {
		jobName = strings.ReplaceAll(jobName, sep, "_")
	}

	// add a suffix to the job name in case the image is being
	// retagged to both public/ and unlisted/ mcr repositories
	suffix := strings.Split(retag.Destination, "/")[0]
	return strings.Join([]string{jobName, suffix}, "_")
}
