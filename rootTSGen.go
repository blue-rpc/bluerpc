package bluerpc

import (
	"os"
	"strings"
)

// // procedureInfo is a generic struct that embeds procedureInfoBase.
// // MAYBE this should be different for QUERY and MUTATION since QUERY has no INPUT
// type procedureInfo struct {
// 	query interface{}
// 	input       interface{}
// 	output      interface{}
// }

// // routeNode is a struct to represent a node in the tree.
// type routeNode struct {
// 	Children map[string]*routeNode
// 	query    *procedureInfo
// 	mutation *procedureInfo
// }

// // root is the root node of the tree.
// var root *routeNode

// func init() {
// 	root = &routeNode{
// 		Children: make(map[string]*routeNode),
// 		query:    nil,
// 		mutation: nil,
// 	}
// }

func generateTs(app *App) error {
	builder := strings.Builder{}
	AddRpcFunc(&builder)

	builder.WriteString("export const rpcAPI =")
	nodeToTS(&builder, app.startRoute, true, "")
	builder.WriteString("as const;")

	file, err := os.Create(app.config.OutputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(builder.String())
	return err

}

func AddRpcFunc(builder *strings.Builder) {
	const text = "async function rpcCall<T>(params: { query?: any, input?: any } | undefined, apiRoute: string): Promise<T> {\n" +
		"  if (params === undefined) {\n" +
		"    const res = await fetch(apiRoute)\n" +
		"    const json = await res.json()\n" +
		"    return json as T\n" +
		"  }\n" +
		"  const url = Object.keys(params.query).length === 0 ? apiRoute : `${apiRoute}?${Object.keys(params.query).map(key => `${key}=${params.query[key]}`).join('&')}`\n" +
		"  const res = await fetch(url, {\n" +
		"    body: Object.keys(params.input).length === 0 ? undefined : params.input\n" +
		"  })\n" +
		"  const json = await res.json()\n" +
		"  return json as T\n" +
		"};\n"
	builder.WriteString(text)
}
