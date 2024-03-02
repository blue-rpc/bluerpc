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

		for _, slug := range keys {

			fullPath := currentPath + slug

			proc := router.procedures[slug]

			//this string split handles the case where there this is a nested dynamic route, something like /:id/name
			tsProcPath, err := splitStringOnSlash(slug)

			if err != nil {
				panic(err)
			}

			for _, path := range tsProcPath {

				path = strings.ReplaceAll(path, "/", "")
				if path[0] == ':' {
					path = path[1:]
				}
				stringBuilder.WriteString(fmt.Sprintf("[`%s`]:{", path))

			}
			if proc.method == STATIC {
				continue
			}
			if proc.protected {
				stringBuilder.WriteString("_")
			}
			switch proc.method {
			case QUERY:
				stringBuilder.WriteString("query: async ")
				query, output := proc.querySchema, proc.outputSchema
				genTSFuncFromQuery(stringBuilder, query, output, fullPath, proc.dynamicSlugs)
			case MUTATION:
				stringBuilder.WriteString("mutation: async ")
				query, input, output := proc.querySchema, proc.inputSchema, proc.outputSchema
				genTSFuncFromMutation(stringBuilder, query, input, output, fullPath, proc.dynamicSlugs)

			}

			//if this is a nested route it will be of the form of some nested objects. Thus we need to close each object that we created.
			stringBuilder.WriteString(strings.Repeat("}", len(tsProcPath)))
			stringBuilder.WriteString(",")
		}
	}

	if router.routes != nil {
		keys := getSortedKeys(router.routes)
		for i, path := range keys {
			tsObjectPath := strings.ReplaceAll(path, "/", "")
			if tsObjectPath[0] == ':' {
				tsObjectPath = tsObjectPath[1:]
			}
			stringBuilder.WriteString(fmt.Sprintf("[`%s`]:", tsObjectPath))

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
