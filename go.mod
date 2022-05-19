module go.astrophena.name/exp

go 1.18

require (
	github.com/PuerkitoBio/goquery v1.8.0
	golang.org/x/net v0.0.0-20220121210141-e204ce36a2ba // indirect
)

require github.com/tailscale/sqlite v0.0.0-20220402182010-0300126d72de

require github.com/andybalholm/cascadia v1.3.1 // indirect

replace github.com/tailscale/sqlite => github.com/astrophena/sqlite v0.0.0-20220519145847-5296ae056b4a
