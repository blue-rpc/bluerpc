package bluerpc

import (
	"fmt"
	"os"
	"strings"
)

func generateTs(app *App) error {
	builder := strings.Builder{}
	addRpcFunc(&builder, app)

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

func addRpcFunc(builder *strings.Builder, app *App) {

	var host string
	if app.config.ServerURL != "" {
		host = fmt.Sprintf(`const host = "%s%s";`, app.config.ServerURL, app.port)
	} else {
		host = `const host = "";`
	}

	text := "/* eslint-disable @typescript-eslint/no-explicit-any */\n" +
		"async function rpcCall<T>( apiRoute: string, params?: { query?: any, input?: any } ): Promise<T> {\n" +
		"  if (params === undefined) {\n" +
		"    const res = await fetch(apiRoute)\n" +
		"    const json = await res.json()\n" +
		"    return json as T\n" +
		"  }\n" +
		host + "\n" +
		"let path = apiRoute;\n" +
		"if (Object.keys(params.query).length !== 0){\n" +
		"  path += `?${Object.keys(params.query).map(key => `${key}=${params.query[key]}`).join('&')}`\n" +
		"}\n" +
		"const url = host + path;\n" +
		"  const res = await fetch(url, {\n" +
		"    body: !params.input || Object.keys(params.input).length === 0 ? undefined : params.input\n" +
		"  })\n" +
		"  const json = await res.json()\n" +
		"  return json as T\n" +
		"}\n"
	builder.WriteString(text)
}
