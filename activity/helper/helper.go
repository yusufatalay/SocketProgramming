package helper

import (
	"strconv"
	"strings"
)

func CreateHTTPResponse(servername string, statuscode int, status string, title string, body string) string {
	response := strings.Builder{}
	afterheader := strings.Builder{}
	afterheader.WriteString("<!DOCTYPE html>\n")
	afterheader.WriteString("<html lang=\"en\">\n")
	afterheader.WriteString("<head>\n")
	afterheader.WriteString("<meta charset=\"UTF-8\">\n")
	afterheader.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	afterheader.WriteString("<title>\n")
	afterheader.WriteString(servername)
	afterheader.WriteString("</title>\n")
	afterheader.WriteString("</head>\n")
	afterheader.WriteString("<body>\n")
	afterheader.WriteString("<h1>\n")
	afterheader.WriteString(title)
	afterheader.WriteRune('\n')
	afterheader.WriteString("</h1>\n")
	afterheader.WriteString("<p>\n")
	afterheader.WriteString(body)
	afterheader.WriteRune('\n')
	afterheader.WriteString("</p>\n")
	afterheader.WriteString("</body>\n")

	response.WriteString("HTTP/1.0 ")
	response.WriteString(strconv.Itoa(statuscode))
	response.WriteString(" ")
	response.WriteString(status)
	response.WriteString("\r\n")
	response.WriteString("Content-Type: text/html")
	response.WriteString("\r\n")
	response.WriteString("Content-Length: ")
	response.WriteString(strconv.Itoa(len([]byte(afterheader.String()))))
	response.WriteString("\r\n")
	response.WriteString("\r\n")
	response.WriteString(afterheader.String())

	return response.String()
}

