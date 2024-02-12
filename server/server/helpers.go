package server

import (
	"fmt"
	"strings"

	json "github.com/de-marauder/gojson/gojson"
)

// Compares requested route with registered route to see if they match
func performRoutePatternMatch(req *HTTPRequest, routeMapVal string, p Params) (string, bool) {
	reqR := strings.TrimRight(req.Route, "/")
	method := strings.Split(routeMapVal, "-")[0]
	r := strings.Split(routeMapVal, "-")[1]
	if method != req.Method {
		return "", false
	}

	reqRSl := strings.Split(reqR, "/")
	rSl := strings.Split(r, "/")

	if len(reqRSl) != len(rSl) {
		return "", false
	}

	for i := range reqRSl {
		vR, vr := reqRSl[i], rSl[i]
		if isParamToken(vr) {
			p[strings.TrimLeft(vr, ":")] = vR
		} else if vR == vr {
			continue
		} else {
			return "", false
		}
	}
	return r, true
}

// Check if a route element is tokenized (i.e. prefixed with a ":")
func isParamToken(s string) bool {
	if s != "" {
		return strings.Split(s, "")[0] == ":"
	} else {
		return false
	}
}

// Set default response headers
func setResHeaders(res *HTTPResponse) {
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	headers["Server"] = "go-express"
	res.Headers = headers
}

// Convert HTTP request to a more usable struct form
// Struct type is HTTPRequest
func parseReqToStruct(message string) *HTTPRequest {
	msgSlice := strings.Split(message, "\r")

	var (
		scheme    string
		fullRoute string
		route     string
		query     Query
		method    string
		version   string
		headers   Headers
		body      string
	)

	counter := 0
	headers = make(map[string]string)
	for lineNo, content := range msgSlice {
		// Read details from scheme (first line of HTTP request)
		if lineNo == 0 {
			scheme = content
			schemeSlice := strings.Fields(scheme)
			method = schemeSlice[0]
			fullRoute = schemeSlice[1]
			version = schemeSlice[2]
			counter++
			continue
		}
		// build headers
		contentSlice := strings.Split(content, ": ")

		// end loop after headers or if a header line cannot be split into a key value pair
		if content == "\r\n" || len(contentSlice) != 2 {
			break
		}
		headers[contentSlice[0]] = contentSlice[1]
		counter += 1
	}

	// Extract route and query
	fullRouteSlice := strings.Split(fullRoute, "?")
	route = fullRouteSlice[0]
	if len(fullRouteSlice) == 2 {
		if len(route) > 1 {
			route = strings.TrimRight(route, "/")
		}
		query = parseQueryToMap(fullRouteSlice[1])
	}

	// build body
	body = strings.Join(msgSlice[counter+1:], "\r\n")

	req := NewHTTPRequest()
	req.Version = version
	req.Route = route
	req.Query = query
	req.Method = method
	req.Headers = headers
	req.Body = json.MustParse(body)

	return req
}

// Convert struct type HTTPResponse to raw HTTP response string
func parseResStructToRaw(res *HTTPResponse) string {
	response := parseResponseStatusLine(res) + "\r\n" + parseHeadersToString(res.Headers) + "\r\n" + parseBody(res) + "\r\n"
	return response
}

// Convert map Headers to a string
func parseHeadersToString(headers Headers) string {
	var parsedHeaders string
	for key, val := range headers {
		parsedHeaders += key + ": " + val + "\r\n"
	}
	return parsedHeaders
}

// Builds the first line in the raw HTTP response (the status line)
func parseResponseStatusLine(res *HTTPResponse) string {
	resStatusLine := fmt.Sprint(res.Version) + " " + fmt.Sprint(res.StatusCode) + " " + HTTPStatusCodeMap[res.StatusCode]
	return resStatusLine
}

// convert response body interface to string
func parseBody(res *HTTPResponse) string {
	body := fmt.Sprint(res.Body)
	res.Headers["Content-Length"] = fmt.Sprint(len(body))
	return body
}

// Stringify JSON
func parseJsonToString(json map[string]string) string {
	var j string = "\"{"

	counter := 1
	for key, val := range json {
		if counter == len(json) {
			j += fmt.Sprintf("\\\"%v\\\":\\\"%v\\\"", key, val)
		} else {
			j += fmt.Sprintf("\\\"%v\\\":\\\"%v\\\",", key, val)
		}
		counter++
	}
	j += "}\""
	return j
}

// Convert query string from the full route to a key value map
func parseQueryToMap(q string) Query {
	qSlice := strings.Split(q, "&")
	query := make(Query)
	for _, qu := range qSlice {
		quSlice := strings.Split(qu, "=")
		key := quSlice[0]
		val := ""
		if len(quSlice) == 2 {
			val = quSlice[1]
		}
		query[key] = val
	}
	return query
}

func concatenateAllHandlers(s *Server, mp routeMapValue) []HTTPRequestHandler {
	m := s.middlewares[:mp.midx]
	middlewareHandlers := make([][]HTTPRequestHandler, len(m))
	copy(middlewareHandlers, m)
	allHandlers := flatten2DSlice(append(middlewareHandlers, mp.handlers))
	return allHandlers
}

func flatten2DSlice[T any](slice [][]T) []T {
	var result []T
	for _, row := range slice {
		result = append(result, row...)
	}
	return result
}
