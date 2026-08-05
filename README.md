![EthPhish logo](static/images/gophish_purple.png)

<div align="center">
<h1>EthPhish</h1>
</div>

EthPhish é um fork corporativo do Anglerphish 1.3.0 destinado exclusivamente a
simulações éticas, autorizadas e mensuráveis de conscientização. O baseline
herdado permanece rastreável, enquanto a evolução do produto prioriza
segregação multitenant, privacidade e operação segura.

See also the Medium [article](https://medium.com/@gpetro/anglerphish-6dc3e5520242).

---

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Features and Enhancements](#features-and-enhancements)
- [Visual Previews](#visual-previews)
- [Contributors](#contributors)
- [A fork based on original Gophish v0.12.1:](#a-fork-based-on-original-gophish-v0121)
  - [Gophish: Open-Source Phishing Toolkit](#gophish-open-source-phishing-toolkit)
  - [Install](#install)
  - [Building From Source](#building-from-source)
  - [Setup](#setup)
  - [Documentation](#documentation)
  - [Issues](#issues)
  - [License](#license)

---

## Features and Enhancements

See [FEATURES.md](FEATURES.md) for the full list of features and enhancements.

For the latest changes check out [CHANGELOG.md](CHANGELOG.md).

## Visual Previews

![1](static/images/1.gif)
![2](static/images/2.gif)
![3](static/images/3.gif)
![4](static/images/4.gif)

## Contributors

[![Contributors](https://contrib.rocks/image?repo=geopetro/anglerphish)](https://github.com/geopetro/anglerphish/graphs/contributors)

Includes everyone across the project's full history — the original Gophish authors inherited through the fork as well as Anglerphish contributors.

## A fork based on original Gophish v0.12.1:

![Build Status](https://github.com/geopetro/anglerphish/workflows/CI/badge.svg) [![GoDoc](https://godoc.org/github.com/gophish/gophish?status.svg)](https://godoc.org/github.com/gophish/gophish)

### Gophish: Open-Source Phishing Toolkit

[Gophish](https://getgophish.com) is an open-source phishing toolkit designed for businesses and penetration testers. It provides the ability to quickly and easily setup and execute phishing engagements and security awareness training.

### Install

Installation of Anglerphish remains dead-simple - just download and extract the zip containing the [release for your system](https://github.com/geopetro/anglerphish/releases/), and run the binary. Anglerphish has also binary releases for Windows, Mac, and Linux platforms.

### Building From Source

Para compilar o EthPhish, instale Go 1.24.5 e um compilador C, então execute
`CGO_ENABLED=1 go build -o ethphish .`. Consulte o runbook de desenvolvimento
local para os pré-requisitos e a execução via Docker Compose.

### Setup
No ambiente local, inicie com `docker compose up -d` e acesse a superfície web
em `https://localhost:9443`. O painel administrativo não é publicado no host.
Na primeira execução, a credencial administrativa temporária é gerada de forma
aleatória e registrada apenas nos logs do servidor local; altere-a no primeiro
acesso. Consulte o runbook de desenvolvimento local para a configuração e a
recuperação PostgreSQL.

### Documentation

Documentation for Anglerphish - Documentation section includes several Anglerphish additions such as newly added API Endpoints.

Documentation of the original gophish can be found on the official [site](http://getgophish.com/documentation).

### Issues

🐞 Found a bug? Feel free to [file an issue](https://github.com/geopetro/anglerphish/releases/issues/new) — feedback is always welcome!

### License
```
MIT License

Copyright (c) 2013–2020 Jordan Wright
Copyright (c) 2025–2026 George Petropoulos

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

----------------------------------------------------------------
Fork Attribution
----------------------------------------------------------------

Anglerphish is an enhanced fork of Gophish v0.12.1,
originally created by Jordan Wright.

----------------------------------------------------------------
Intended Use Notice (Non-Binding Advisory)
----------------------------------------------------------------

Anglerphish is intended exclusively for authorized security
testing, phishing simulations, user awareness training,
and defensive cybersecurity research.

Users are responsible for ensuring compliance with all
applicable laws and obtaining proper authorization before use.

This notice does not modify or supersede the terms of the MIT License.
```
