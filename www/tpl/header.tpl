<!DOCTYPE html>
<head>
<meta name="Robots" content="index,follow">
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link rel="icon" href="/style/icon.png" type="image/png">
<link rel="stylesheet" type="text/css" href="/style/style.css">
{{ if eq .Template "query.tpl" }}<link href="{{.PfxPath}}/{{.BasePath}}/rss" type="application/rss+xml" rel="alternate" title="{{.Sysname}} {{.BasePath}} :: RSS feed" />{{ end }}
{{ if eq .Template "blog.tpl" }}<link href="{{.PfxPath}}/{{.BasePath}}+topics/rss" type="application/rss+xml" rel="alternate" title="{{.Sysname}} {{.BasePath}} :: RSS feed" />{{ end }}

<title>{{.Sysname}}</title>
</head>
<body>
<div id="body">
<table id="header">
  <tr>
    <td class="title">
      <span class="logo"><a href="/"><img class="logo" src="/style/icon.png">{{.Sysname}}</a></span>
{{ if eq .BasePath "" }}
      <span class="info">II/IDEC networks :: <a href="{{ $.PfxPath }}/echo/all">New posts</a>
{{ else if gt (len .Topics) 0}}
      <span class="info">II/IDEC networks {{ with .Echo }} :: <a href="{{$.PfxPath}}/echo/{{.}}">{{.}}</a> :: <span class="info">{{index $.Echolist.Info .}}</span>{{end}}
{{ else }}
      {{ if eq .Template "query.tpl" }}
      <span class="info">II/IDEC networks {{ with .Echo }} :: <a href="{{$.PfxPath}}/{{.}}">{{.}}</a> :: <span class="info">{{index $.Echolist.Info .}} / feed</span>{{end}}
      {{ else }}
      <span class="info">II/IDEC networks {{ with .Echo }} :: <a href="{{$.PfxPath}}/{{.}}">{{.}}</a> :: <span class="info">{{index $.Echolist.Info .}}</span>{{end}}
      {{ end }}
{{ end }}
</span>
    </td>
    <td class="links">
      <span>
      {{ template "links.tpl" }}
      {{ if .User.Name }}
      {{ if eq .BasePath "profile" }}
      <a href="/logout">Logout</a>
      {{ else }}
      <a href="/profile">{{.User.Name}}</a>
      {{ end }}

      {{ with .Echo }}
      {{ if $.Topic }}
      :: <a href="{{$.PfxPath}}/{{$.Topic}}/reply/new">New</a>
      {{ else }}
      :: <a href="{{$.PfxPath}}/{{.}}/new">New</a>
      {{ end }}
      {{ end }}

      {{ else if eq .BasePath "login" }}
      <a href="/register">Register</a>
      {{ else }}
      <a href="/login">Login</a>
      {{ end }}

      </span>
    </td>
</tr>
</table>
