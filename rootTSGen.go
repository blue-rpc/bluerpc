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
		"type Method = \"GET\" | \"POST\"\n" +
		"async function rpcCall<T>(\n" +
		"  apiRoute: string,\n" +
		"  method: Method,\n" +
		"  params?: { query?: any; input?: any },\n" +
		"  headers?: HeadersInit\n" +
		"): Promise<{ body: T; status: number; headers: Headers }> {\n" +
		"  const requestOptions: RequestInit = {\n" +
		"    method: method,\n" +
		"    headers: headers,\n" +
		"  };\n" +
		"  if (params?.input) {\n" +
		"    requestOptions.body = JSON.stringify(params.input);\n" +
		"  }\n" +
		host + "\n" +
		"  let path = apiRoute;\n" +
		"  if (params?.query && Object.keys(params.query).length !== 0) {\n" +
		"    path += `?${Object.keys(params.query)\n" +
		"      .map(key => {\n" +
		"        if (key.includes('Slug')) {\n" +
		"          return '';\n" +
		"        }\n" +
		"        return `${encodeURIComponent(key)}=${encodeURIComponent(params.query[key])}`;\n" +
		"      })\n" +
		"      .join('&')}`;\n" +
		"  }\n" +
		"  const url = host + path\n" +
		"  const res = await fetch(url, requestOptions);\n" +
		"  const contentType = res.headers.get('content-type');\n" +
		"  let body: any;\n" +
		"  if (contentType?.includes('application/json')) {\n" +
		"    body = await res.json();\n" +
		"  } else if (contentType?.includes('text')) {\n" +
		"    body = await res.text();\n" +
		"  } else {\n" +
		"    body = await res.blob(); // or arrayBuffer, depending on the expected response\n" +
		"  }\n" +
		"\n" +
		"  return { \n" +
		"    body: body as T, \n" +
		"    status: res.status, \n" +
		"    headers: res.headers \n" +
		"  };\n" +
		"}\n"
	builder.WriteString(text)
}
