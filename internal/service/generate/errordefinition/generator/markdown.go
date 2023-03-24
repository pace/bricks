package generator

import (
	"encoding/json"
	"fmt"
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

func (g *Generator) BuildMarkdown(source string) (string, error) {
	data, err := loadDefinitionData(source)
	if err != nil {
		return "", err
	}

	// parse definitions
	eds, err := g.parseDefinitions(data)
	if err != nil {
		return "", err
	}

	return g.generateMarkdown(eds)
}

func (g *Generator) parseDefinitions(data []byte) (ErrorDefinitions, error) {
	var parsedData []ErrorDefinition
	err := json.Unmarshal(data, &parsedData)
	if err != nil {
		return nil, err
	}

	eds := make(ErrorDefinitions)
	for _, ed := range parsedData {
		if _, ok := eds[ed.Service]; !ok {
			eds[ed.Service] = make(map[string][]GroupedErrorDetails)
		}

		eds[ed.Service][ed.Status] = append(eds[ed.Service][ed.Status], GroupedErrorDetails{
			Code:  ed.Code,
			Title: ed.Title,
		})
	}

	return eds, nil
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
