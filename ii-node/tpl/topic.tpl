{{template "header.tpl"}}

<div class="topic">
{{ range .Msg }}
<div class="msg">
<span class="msgsubj">{{.Subj}}</span>
<br/>
<span class="msginfo">{{.From}}({{.Addr}}) -> {{.To}}</span>
<br/>
<span class="msgtext">
{{with .Text}}
{{call $.Render .}}
{{end}}
</span>
</div>
{{ end }}
</div>

{{template "footer.tpl"}}
