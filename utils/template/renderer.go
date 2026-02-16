package template

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type Renderer struct {
	template  string
	variables map[string]interface{}
}

func NewRenderer(template string) *Renderer {
	return &Renderer{
		template:  template,
		variables: make(map[string]interface{}),
	}
}

func (r *Renderer) SetVariable(key string, value interface{}) {
	r.variables[key] = value
}

func (r *Renderer) SetVariables(variables map[string]interface{}) {
	for k, v := range variables {
		r.variables[k] = v
	}
}

func (r *Renderer) Render() (string, error) {
	re := regexp.MustCompile(`\[(\w+)\]`)
	result := re.ReplaceAllStringFunc(r.template, func(match string) string {
		key := strings.Trim(match, "[]")
		value, exists := r.variables[key]
		if !exists {
			return match
		}
		return fmt.Sprintf("%v", value)
	})
	return result, nil
}

func (r *Renderer) RenderWithMap(variables map[string]interface{}) (string, error) {
	r.SetVariables(variables)
	return r.Render()
}

func GetVariablesFromTemplate(template string) []string {
	re := regexp.MustCompile(`\[(\w+)\]`)
	matches := re.FindAllStringSubmatch(template, -1)

	variablesMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			variablesMap[match[1]] = true
		}
	}

	variables := make([]string, 0, len(variablesMap))
	for v := range variablesMap {
		variables = append(variables, v)
	}

	return variables
}

func ParseVariablesJSON(jsonStr string) ([]string, error) {
	if jsonStr == "" {
		return []string{}, nil
	}

	var variables []string
	err := json.Unmarshal([]byte(jsonStr), &variables)
	if err != nil {
		return nil, fmt.Errorf("failed to parse variables JSON: %w", err)
	}

	return variables, nil
}

func VariablesToJSON(variables []string) (string, error) {
	if len(variables) == 0 {
		return "[]", nil
	}

	jsonBytes, err := json.Marshal(variables)
	if err != nil {
		return "", fmt.Errorf("failed to marshal variables to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

func ValidateTemplateVariables(template string, requiredVariables []string) ([]string, error) {
	templateVars := GetVariablesFromTemplate(template)

	requiredMap := make(map[string]bool)
	for _, v := range requiredVariables {
		requiredMap[v] = true
	}

	var missing []string
	for _, requiredVar := range requiredVariables {
		found := false
		for _, templateVar := range templateVars {
			if templateVar == requiredVar {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, requiredVar)
		}
	}

	return missing, nil
}
