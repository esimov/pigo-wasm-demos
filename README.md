# pigo-wasm-demos

[![License](https://img.shields.io/github/license/esimov/pigo-wasm-demos)](https://github.com/esimov/pigo-wasm-demos/blob/master/LICENSE)

<p align="center"><img src="https://user-images.githubusercontent.com/883386/80915158-06911a80-8d59-11ea-93bd-eca98750ad62.png" alt="Pigo Wasm demos" title="Pigo Wasm demos" width="400"/></p>

This repository is a collection of real time Webassembly demos based on the [Pigo](https://github.com/esimov/pigo) face detection library, showing various use cases the library can be used for. **It is continuously updated**.

## Install
**Notice: at least Go 1.13 is required prior running the demos!**

```bash
$ go get -u -v github.com/esimov/pigo-wasm-demos 

```

## Run
To run the demos is as simple as typing a single line of command. This will build the package and produce an executable WebAssembly file which can be served over a http server. A new tab will be opened in the user's default browser.

## Demos

### Masquerade
```bash
$ make demo1
```
![pigo_wasm_masquarade](https://user-images.githubusercontent.com/883386/82048111-ae450b80-96bc-11ea-9f22-7039ce937140.gif)


#### Key bindings:
<kbd>q</kbd> - Show/hide face marker<br/>
<kbd>z</kbd> - Show/hide pupils<br/>
<kbd>w</kbd> - Show/hide eye mask<br/>
<kbd>s</kbd> - Show/hide mouth mask<br/>
<kbd>e</kbd> - Select the next eye mask<br/>
<kbd>d</kbd> - Select the previous eye mask<br/>
<kbd>r</kbd> - Select the next mouth mask<br/>
<kbd>f</kbd> - Select the previous mouth mask<br/>
<kbd>x</kbd> - Show the detected face coordinates<br/>

### Faceblur
```bash
$ make demo2
```
![pigo_wasm_faceblur](https://user-images.githubusercontent.com/883386/82048882-16482180-96be-11ea-9246-836c378b7eb7.gif)


### Pixelate
```bash
$ make demo3
```
![pigo_wasm_pixelate](https://user-images.githubusercontent.com/883386/82049123-80f95d00-96be-11ea-801d-6e5a50d36114.gif)

### Face triangulator
```bash
$ make demo4
```
![pigo_wasm_triangulate](https://user-images.githubusercontent.com/883386/82050510-ebab9800-96c0-11ea-84fb-00475076d33f.gif)


## Author

* Endre Simo ([@simo_endre](https://twitter.com/simo_endre))

## License

Copyright © 2020 Endre Simo

This software is distributed under the MIT license. See the [LICENSE](https://github.com/esimov/pigo-wasm-demos/blob/master/LICENSE) file for the full license text.
