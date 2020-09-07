{{template "header.tpl" $}}
{{template "pager.tpl" $}}
<a class="rss" href="/{{.BasePath}}/rss">RSS</a>
<div id="topic">
{{ range .Msg }}
<div class="msg">
{{ if has_avatar .From }}
<img class="avatar" src="/avatar/{{.From}}">
{{ end }}
<a class="msgid" href="/{{.MsgId}}#{{.MsgId}}">#</a><span class="subj"> <a href="/{{. | repto}}#{{. | repto}}">{{with .Subj}}{{.}}{{else}}No subject{{end}}</a></span><br>
<span class="echo"><a href="/{{.Echo}}">{{.Echo}}</a></span><br>
<span class="info">{{.From}}({{.Addr}}) &mdash; {{.To}}<br>{{.Date | fdate}}</span><br>
<div class="text">
<br>
{{with .Text}}
{{. | msg_format}}
{{end}}
<br>
{{if $.User.Name}}
<span class="reply"><a href="/{{.MsgId}}/reply">Reply</a></span>
{{end}}
{{ if msg_access . $.User }}
 :: <span class="reply"><a href="/{{.MsgId}}/edit">Edit</a></span>
{{ end }}
{{if $.User.Name}}
<br>
{{end}}

</div>
</div>
{{ end }}
</div>
{{template "pager.tpl" $}}

{{template "footer.tpl"}}
