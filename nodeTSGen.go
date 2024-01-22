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
			fullPath := currentPath + slug
			fmt.Println("full path from node to ts", fullPath)

			proc := router.procedures[slug]
			slug = strings.Replace(slug, "/", "", 1)
			//this string split handles the case where there this is a nested dynamic route, something like /:id/name
			tsProcPath := strings.Split(slug, "/")

			for _, path := range tsProcPath {
				path = strings.ReplaceAll(path, "/", "")
				if path[0] == ':' {
					stringBuilder.WriteString(fmt.Sprintf("[`%s`]:{", path[1:]))
				} else {
					stringBuilder.WriteString(fmt.Sprintf("%s:{", path))
				}
			}

			if proc.method == QUERY {
				stringBuilder.WriteString("query: async ")
				query, output := proc.querySchema, proc.outputSchema
				genTSFuncFromQuery(stringBuilder, query, output, fullPath, proc.dynamicSlugs)
			}

			if proc.method == MUTATION {
				stringBuilder.WriteString("mutation: async ")
				query, input, output := proc.querySchema, proc.inputSchema, proc.outputSchema
				genTSFuncFromMutation(stringBuilder, query, input, output, fullPath, proc.dynamicSlugs)
			}

			//if this is a nested route it will be of the form of some nested objects. Thus we need to close each object that we created.
			stringBuilder.WriteString(strings.Repeat("}", len(tsProcPath)))
			if i != len(keys)-1 {
				stringBuilder.WriteString(",")
			}
		}
	}

	if router.routes != nil {
		keys := getSortedKeys(router.routes)
		for i, path := range keys {
			tsObjectPath := strings.ReplaceAll(path, "/", "")
			if tsObjectPath[0] == ':' {
				stringBuilder.WriteString(fmt.Sprintf("[`%s`]:", tsObjectPath[1:]))
			} else {
				stringBuilder.WriteString(fmt.Sprintf("%s:", tsObjectPath))
			}

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
