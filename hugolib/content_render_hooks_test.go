// Copyright 2019 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless requiredF by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hugolib

import (
	"testing"
)

func TestRenderHooksRSS(t *testing.T) {
	files := `
-- hugo.toml --
baseURL = "https://example.org"
disableKinds = ["taxonomy", "term"]
-- layouts/index.html --
{{ $p := site.GetPage "p1.md" }}
{{ $p2 := site.GetPage "p2.md" }}
P1: {{ $p.Content }}
P2: {{ $p2.Content }}
-- layouts/index.xml --
{{ $p2 := site.GetPage "p2.md" }}
{{ $p3 := site.GetPage "p3.md" }}
P2: {{ $p2.Content }}
P3: {{ $p3.Content }}
-- layouts/_default/_markup/render-link.html --
html-link: {{ .Destination | safeURL }}|
-- layouts/_default/_markup/render-link.rss.xml --
xml-link: {{ .Destination | safeURL }}|
-- layouts/_default/_markup/render-heading.html --
html-heading: {{ .Text }}|
-- layouts/_default/_markup/render-heading.rss.xml --
xml-heading: {{ .Text }}|
-- content/p1.md --
---
title: "p1"
---
P1. [I'm an inline-style link](https://www.gohugo.io)

# Heading in p1

-- content/p2.md --
---
title: "p2"
---
P2. [I'm an inline-style link](https://www.bep.is)

# Heading in p2

-- content/p3.md --
---
title: "p3"
outputs: ["rss"]
---
P3. [I'm an inline-style link](https://www.example.org)
`
	b := Test(t, files)

	b.AssertFileContent("public/index.html", `
P1: <p>P1. html-link: https://www.gohugo.io|</p>
html-heading: Heading in p1|
html-heading: Heading in p2|
`)
	b.AssertFileContent("public/index.xml", `
P2: <p>P2. xml-link: https://www.bep.is|</p>
P3: <p>P3. xml-link: https://www.example.org|</p>
xml-heading: Heading in p2|
`)
}

// https://github.com/gohugoio/hugo/issues/6629
func TestRenderLinkWithMarkupInText(t *testing.T) {
	b := newTestSitesBuilder(t)
	b.WithConfigFile("toml", `

baseURL="https://example.org"

[markup]
  [markup.goldmark]
    [markup.goldmark.renderer]
      unsafe = true
    
`)

	b.WithTemplates("index.html", `
{{ $p := site.GetPage "p1.md" }}
P1: {{ $p.Content }}

	`,
		"_default/_markup/render-link.html", `html-link: {{ .Destination | safeURL }}|Text: {{ .Text | safeHTML }}|Plain: {{ .PlainText | safeHTML }}`,
		"_default/_markup/render-image.html", `html-image: {{ .Destination | safeURL }}|Text: {{ .Text | safeHTML }}|Plain: {{ .PlainText | safeHTML }}`,
	)

	b.WithContent("p1.md", `---
title: "p1"
---

START: [**should be bold**](https://gohugo.io)END

Some regular **markup**.

Image:

![Hello<br> Goodbye](image.jpg)END

`)

	b.Build(BuildCfg{})

	b.AssertFileContent("public/index.html", `
  P1: <p>START: html-link: https://gohugo.io|Text: <strong>should be bold</strong>|Plain: should be boldEND</p>
<p>Some regular <strong>markup</strong>.</p>
<p>html-image: image.jpg|Text: Hello<br> Goodbye|Plain: Hello GoodbyeEND</p>
`)
}

func TestRenderHookContentFragmentsOnSelf(t *testing.T) {
	files := `
-- hugo.toml --
baseURL = "https://example.org"
disableKinds = ["taxonomy", "term", "RSS", "sitemap", "robotsTXT"]
-- content/p1.md --
---
title: "p1"
---

## A {#z}
## B
## C

-- content/p2.md --
---
title: "p2"
---

## D
## E
## F

-- layouts/_default/_markup/render-heading.html --
Heading: {{ .Text }}|
{{ with .Page }}
Self Fragments: {{ .Fragments.Identifiers }}|
{{ end }}
{{ with (site.GetPage "p1.md") }}
P1 Fragments: {{ .Fragments.Identifiers }}|
{{ end }}
-- layouts/_default/single.html --
{{ .Content}}
`

	b := Test(t, files)

	b.AssertFileContent("public/p1/index.html", `
Self Fragments: [b c z]
P1 Fragments: [b c z]
	`)
	b.AssertFileContent("public/p2/index.html", `
Self Fragments: [d e f]
P1 Fragments: [b c z]
	`)
}
