package bluerpc

import (
	"fmt"
	"sort"
	"strings"
)

func nodeToTS(stringBuilder *strings.Builder, router *Router, isLast bool, currentPath string) {

	stringBuilder.WriteString("{")

	if router.procedures != nil {
		keys := getSortedKeys(router.procedures)
		for i, slug := range keys {
			proc := router.procedures[slug]

			stringBuilder.WriteString(fmt.Sprintf("%s:{", strings.ReplaceAll(slug, "/", "")))
			if proc.method == QUERY {
				stringBuilder.WriteString("query: async ")
				queryParams, output := proc.querySchema, proc.outputSchema
				genTSFuncFromQuery(stringBuilder, queryParams, output, currentPath)
			}

			if proc.method == MUTATION {
				stringBuilder.WriteString("mutation: async ")
				queryParams, input, output := proc.querySchema, proc.inputSchema, proc.outputSchema
				genTSFuncFromMutation(stringBuilder, queryParams, input, output, currentPath)
			}

			stringBuilder.WriteString("}")
			if i != len(keys)-1 {
				stringBuilder.WriteString(",")
			}
		}
	}

	if router.Routes != nil {
		keys := getSortedKeys(router.Routes)
		for i, path := range keys {
			stringBuilder.WriteString(fmt.Sprintf("%s:", strings.ReplaceAll(path, "/", "")))
			nodeToTS(stringBuilder, router.Routes[path], i == len(keys)-1, currentPath+path)
		}

	}

	stringBuilder.WriteString("}")
	if !isLast {
		stringBuilder.WriteString(",")
	}
}
func getSortedKeys[values any](m map[string]values) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
