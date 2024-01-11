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
				query, output := proc.querySchema, proc.outputSchema
				genTSFuncFromQuery(stringBuilder, query, output, currentPath)
			}

			if proc.method == MUTATION {
				stringBuilder.WriteString("mutation: async ")
				query, input, output := proc.querySchema, proc.inputSchema, proc.outputSchema
				genTSFuncFromMutation(stringBuilder, query, input, output, currentPath)
			}

			stringBuilder.WriteString("}")
			if i != len(keys)-1 {
				stringBuilder.WriteString(",")
			}
		}
	}

	if router.routes != nil {
		keys := getSortedKeys(router.routes)
		for i, path := range keys {
			stringBuilder.WriteString(fmt.Sprintf("%s:", strings.ReplaceAll(path, "/", "")))
			nodeToTS(stringBuilder, router.routes[path], i == len(keys)-1, currentPath+path)
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
