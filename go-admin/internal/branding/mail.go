package branding

import (
	"bytes"
	"encoding/base64"
	"html/template"
)

const emailLogoSize = 96

// EmailLogoBase64 returns the base64-encoded PNG data URI for an email logo.
// text is the first character of the app name; bgHex is the brand color.
func EmailLogoBase64(text, bgHex string) (string, error) {
	if text == "" {
		text = "S"
	}
	if bgHex == "" {
		bgHex = "#1d4ed8"
	}
	ch := []rune(text)[0:1]
	data, err := RenderDefaultLogoPNG(string(ch), bgHex, emailLogoSize)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data), nil
}

// EmailTemplateData holds variables for rendering the email template.
type EmailTemplateData struct {
	LogoDataURI template.URL // safe data URI for <img src="...">
	AppName     string
	Title       string
	Body        template.HTML
	Footer      string
}

var emailTmpl = template.Must(template.New("email").Parse(emailTemplateHTML))

// RenderEmailTemplate renders an HTML email with an inline base64 logo.
// It calls RenderDefaultLogoPNG internally to generate the logo PNG.
func RenderEmailTemplate(appName, title, body, footer, bgHex string) (string, error) {
	logoURI, err := EmailLogoBase64(firstChar(appName), bgHex)
	if err != nil {
		return "", err
	}

	data := EmailTemplateData{
		LogoDataURI: template.URL(logoURI),
		AppName:     appName,
		Title:       title,
		Body:        template.HTML(body),
		Footer:      footer,
	}

	var buf bytes.Buffer
	if err = emailTmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func firstChar(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return "S"
	}
	return string(runes[0])
}

const emailTemplateHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>{{.Title}}</title>
</head>
<body style="margin:0;padding:0;background:#f4f4f5;font-family:sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f5;padding:40px 0;">
    <tr>
      <td align="center">
        <table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:8px;overflow:hidden;box-shadow:0 2px 8px rgba(0,0,0,.06);">
          <!-- Header -->
          <tr>
            <td style="padding:24px 32px;border-bottom:1px solid #e5e7eb;">
              <img src="{{.LogoDataURI}}" width="96" height="96" alt="{{.AppName}} logo"
                   style="display:block;border-radius:12px;">
              <span style="display:block;margin-top:8px;font-size:18px;font-weight:600;color:#111827;">{{.AppName}}</span>
            </td>
          </tr>
          <!-- Body -->
          <tr>
            <td style="padding:32px;color:#374151;font-size:15px;line-height:1.7;">
              <h2 style="margin:0 0 16px;font-size:20px;color:#111827;">{{.Title}}</h2>
              {{.Body}}
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="padding:16px 32px;border-top:1px solid #e5e7eb;color:#6b7280;font-size:12px;">
              {{.Footer}}
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`
