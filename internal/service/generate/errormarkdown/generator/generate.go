package generator

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
)

type ErrorDefinition struct {
	Service string
	Status  string
	Code    string
	Title   string
}

type ErrorDefinitions map[string]map[string][]GroupedErrorDetails

type GroupedErrorDetails struct {
	Code  string
	Title string
}

type Generator struct{}

func (g *Generator) Generate(in io.Reader) (string, error) {
	eds := g.parseInput(in)
	return g.generateMarkdown(eds)
}

func (g *Generator) parseInput(r io.Reader) ErrorDefinitions {
	dec := yaml.NewDecoder(r)

	eds := make(ErrorDefinitions)
	for {
		var ed ErrorDefinition
		if dec.Decode(&ed) != nil {
			break
		}

		if _, ok := eds[ed.Service]; ok {
			eds[ed.Service][ed.Status] = append(eds[ed.Service][ed.Status], GroupedErrorDetails{
				Code:  ed.Code,
				Title: ed.Title,
			})
		} else {
			eds[ed.Service] = make(map[string][]GroupedErrorDetails)
		}
	}

	return eds
}

func (g *Generator) generateMarkdown(eds ErrorDefinitions) (string, error) {
	output := &strings.Builder{}
	for service, statuses := range eds {
		_, err := output.WriteString(fmt.Sprintf("# %s\n", service))
		if err != nil {
			return "", err
		}

		for status, details := range statuses {
			_, err := output.WriteString(fmt.Sprintf("## %s\n", status))
			if err != nil {
				return "", err
			}
			_, err = output.WriteString(fmt.Sprint(`|Code|Title|
|-----------|-----------|
`))
			if err != nil {
				panic(err)
			}
			for _, detail := range details {
				_, err := output.WriteString(fmt.Sprintf("|%s|%s|\n", detail.Code, detail.Title))
				if err != nil {
					return "", err
				}
			}
		}
	}

	return output.String(), nil
}
