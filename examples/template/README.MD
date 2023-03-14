# Template generation example
Example of build process for html template where all resources are embedded as base64 strings.
## Requirements:
- [parcel](https://parceljs.org/)
## How to build template:
`$ npx parcel build src/index.html`

Output file: `dist/index.html`
## How to embed resources:
Use `data-url:` [scheme](https://parceljs.org/features/bundle-inlining/#inlining-as-a-data-url).

Examples:

- html: `<img src="data-url:./img/image.png">`

- css: `<background-image: url(data-url:./img/bg.svg);>`