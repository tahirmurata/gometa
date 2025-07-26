/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"os"
	"slices"

	"github.com/alecthomas/kong"
	"github.com/tdewolff/minify/v2/minify"
)

var CLI struct {
	Init struct {
	} `cmd:"" help:"Initialize config file."`

	Build struct {
		Domain string `arg:"" help:"The domain."`
	} `cmd:"" help:"Build the html files."`
}

func main() {
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "init":
		if _, err := os.Stat("gometa.csv"); !os.IsNotExist(err) {
			panic("Config file already exists")
		}

		f, err := os.Create("gometa.csv")
		if err != nil {
			panic(fmt.Errorf("Could not create config file: %w", err))
		}
		defer f.Close()

		cw := csv.NewWriter(f)
		cw.Write([]string{"package", "vcs", "repo"})
		cw.Write([]string{"gometa", "git", "https://github.com/tahirmurata/gometa"})
		cw.Flush()

	case "build <domain>":
		err := os.RemoveAll("dist")
		if err != nil {
			panic(fmt.Errorf("Failed to remove all dist folder: %w", err))
		}

		if _, err := os.Stat("gometa.csv"); os.IsNotExist(err) {
			panic("Config file does not exist")
		}

		err = os.Mkdir("dist", 0755)
		if err != nil {
			panic(fmt.Errorf("Failed to make dist dir: %w", err))
		}

		f, err := os.Open("gometa.csv")
		if err != nil {
			panic(fmt.Errorf("Could not open config file: %w", err))
		}
		defer f.Close()
		cr := csv.NewReader(f)
		s, err := cr.ReadAll()
		if err != nil {
			panic(fmt.Errorf("Failed to read all csv: %w", err))
		}
		s = slices.Delete(s, 0, 1)

		m := minify.Default

		for _, r := range s {
			if len(r) != 3 {
				panic("More or less records")
			}

			h, err := os.Create(fmt.Sprintf("dist/%s.html", r[0]))
			if err != nil {
				panic(fmt.Errorf("Failed to create dir: %w", err))
			}
			defer h.Close()

			data := Data{
				Path: fmt.Sprintf("%s/%s", CLI.Build.Domain, r[0]),
				VCS:  r[1],
				Repo: r[2],
			}

			mr := m.Writer("text/html", h)
			defer mr.Close()

			t := template.Must(template.New("html").Parse(htmlTemplate))
			err = t.ExecuteTemplate(mr, "html", data)
			if err != nil {
				panic(fmt.Errorf("Failed to execute html template: %w", err))
			}

			h.Sync()
		}

	default:
		panic(ctx.Command())
	}
}

type Data struct {
	Path string
	VCS  string
	Repo string
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<meta name="go-import" content="{{.Path}} {{.VCS}} {{.Repo}}" />
		<meta http-equiv="refresh" content="0;url=https://pkg.go.dev/{{.Path}}" />
		<meta name="robots" content="noindex,noarchive" />
		<meta name="generator" content="gometa" />
		<style>
			html,
			:host {
				background-color: hsl(220deg 23% 95%);
				color: hsl(234deg 16% 35%);
				-webkit-text-size-adjust: 100%;
				font-family:
					ui-sans-serif, system-ui, sans-serif, "Apple Color Emoji",
					"Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji";
				font-feature-settings: normal;
				font-variation-settings: normal;
			}

			@media (prefers-color-scheme: dark) {
				html {
					background-color: hsl(240deg 21% 15%);
					color: hsl(226deg 64% 88%);
				}
			}

			a {
				color: inherit;
				-webkit-text-decoration: inherit;
				text-decoration: underline;
			}

			p {
				text-align: center;
				margin-top: 1.44rem;
			}
		</style>
	</head>
	<body>
		<p>Redirecting to <a href="https://pkg.go.dev/{{.Repo}}">Go Packages</a>...</p>
	</body>
</html>
`
