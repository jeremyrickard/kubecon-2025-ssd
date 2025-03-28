package retag

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	generateName             = "generate"
	generateShortDescription = "Generate Github Matrix variable for retagging"
	generateLongDescription  = `Generates the json representation of the provided configuration file(s) which will be set as a variable and used in the matrix declaration.
This allows us to spawn a retag job per repository allowing us to scale the retag workflow horizontally`
)

type MCRConfig struct {
	Registry string `yaml:"registry"`
	Repos    []Repo `yaml:"repos"`
}

type Repo struct {
	Name            string           `yaml:"name"`
	DisplayName     string           `yaml:"displayName"`
	Description     string           `yaml:"description"`
	PublisherConfig *PublisherConfig `yaml:"publisherConfiguration"`
}

type PublisherConfig struct {
	RetagConfig Retag `yaml:"azcu"`
}

type Retag struct {
	Source         string    `yaml:"source"`
	Destination    string    `yaml:"destination"`
	DateAdded      time.Time `yaml:"date_added"`
	Tags           []string  `yaml:"tags"`
	Tool           string    `yaml:"tool"`
	EnableTimeBomb bool      `yaml:"enable_timebomb"`
}

type generateCmd struct {
	configFile []string
	retags     []Retag
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

	_ = generateCmd.MarkFlagRequired("config")
	return generateCmd
}

func (gc *generateCmd) validate(_ *cobra.Command, _ []string) error {
	var retags []Retag
	for _, configFile := range gc.configFile {
		data, err := load(configFile)
		if err != nil {
			return err
		}

		rts, err := parse(data)
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
		tool := retag.Tool
		if tool == "" {
			tool = "az"
		}
		matrix[sanitizeJobName(&retag)] = map[string]string{
			"source":          retag.Source,
			"destination":     retag.Destination,
			"date_added":      retag.DateAdded.Format("2006-01-02"),
			"tags":            strings.Join(retag.Tags, ","),
			"tool":            tool,
			"enable_timebomb": strconv.FormatBool(retag.EnableTimeBomb),
		}
	}

	return matrix
}

// load reads the contents of a file and returns the bytes.
func load(file string) ([]byte, error) {
	filebytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return filebytes, nil
}

// parse parses the yaml data and returns a list of retags.
func parse(filebytes []byte) ([]Retag, error) {
	var config MCRConfig
	err := yaml.Unmarshal(filebytes, &config)
	if err != nil {
		return nil, err
	}

	var rts []Retag
	for _, repo := range config.Repos {
		if repo.PublisherConfig == nil {
			return nil, fmt.Errorf("missing retag config for %s", repo.Name)
		}
		retagConfig := repo.PublisherConfig.RetagConfig
		if retagConfig.Source == "" {
			return nil, fmt.Errorf("source is required for retagging into %s", retagConfig.Source)
		}
		retagConfig.Destination = repo.Name
		if len(retagConfig.Tags) < 1 {
			return nil, fmt.Errorf("at least one tag is required for retagging %s", retagConfig.Source)
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
