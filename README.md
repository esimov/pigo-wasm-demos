# pigo-wasm-demos

[![License](https://img.shields.io/github/license/esimov/pigo-wasm-demos)](https://github.com/esimov/pigo-wasm-demos/blob/master/LICENSE)

<p align="center"><img src="https://user-images.githubusercontent.com/883386/80915158-06911a80-8d59-11ea-93bd-eca98750ad62.png" alt="Pigo Wasm demos" title="Pigo Wasm demos" width="400"/></p>

This repository is a collection of Webassembly demos running in real time and using the [Pigo](https://github.com/esimov/pigo) face detection library, showcasing various potential use cases. **It is continuously updated**.

## Install
**Notice: at least Go 1.13 is required in order to run the demos!**

```bash
$ go get -u -v github.com/esimov/pigo-wasm-demos 

```

## Run
Running the demo is as simple as typing a single line of command. This will build the package and produce an executable WebAssembly file which can be served over a http server. A new tab will be opened automatically in the user's default browser. 

## Demos

### Masquerade
```bash
$ make demo1
```
![pigo_wasm_masquarade](https://user-images.githubusercontent.com/883386/82048111-ae450b80-96bc-11ea-9f22-7039ce937140.gif)


#### Key bindings:
<kbd>q</kbd> - Show/hide detected face marker<br/>
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


#### Key bindings:
<kbd>f</kbd> - Show/hide detected face marker<br/>
<kbd>s</kbd> - Show/hide pupils<br/>
<kbd>b</kbd> - Enable/disable face blur<br/>
<kbd>]</kbd> - Increase the blur radius<br/>
<kbd>[</kbd> - Decrease the blur radius<br/>

### Pixelate
```bash
$ make demo3
```
![pigo_wasm_pixelate](https://user-images.githubusercontent.com/883386/82049123-80f95d00-96be-11ea-801d-6e5a50d36114.gif)

#### Key bindings:
<kbd>f</kbd> - Show/hide detected face marker<br/>
<kbd>s</kbd> - Show/hide pupils<br/>
<kbd>=</kbd> - Increase the number of colors<br/>
<kbd>-</kbd> - Decrease the number of colors<br/>
<kbd>]</kbd> - Increase the cells size<br/>
<kbd>[</kbd> - Decrease the cells size<br/>

### Face triangulator
```bash
$ make demo4
```
![pigo_wasm_triangulate](https://user-images.githubusercontent.com/883386/82050510-ebab9800-96c0-11ea-84fb-00475076d33f.gif)

#### Key bindings:
<kbd>f</kbd> - Show/hide detected face marker<br/>
<kbd>s</kbd> - Show/hide pupils<br/>
<kbd>=</kbd> - Increase the number of triangles<br/>
<kbd>-</kbd> - Decrease the number of triangles<br/>
<kbd>]</kbd> - Increase the threshold<br/>
<kbd>[</kbd> - Decrease the threshold<br/>
<kbd>1</kbd> - Increase the stroke size<br/>
<kbd>0</kbd> - Decrease the stroke size<br/>

### Triangulated facemask
```bash
$ make demo5
```
![facemask](https://user-images.githubusercontent.com/883386/132861943-5f130ec2-dae2-4034-9abd-4c9de0de066c.gif)

This demo is meant to be a proof of concept for an idea of generating personalized triangulated face masks. The orange dot at the bottom of the screen is showing up when the head alignment is the most appropriate for making a screen capture and this is when the head is aligned perpendicular (+/- a predefined threshold). This demo can be expanded way further.

#### Key bindings:
<kbd>f</kbd> - Show/hide detected face marker<br/>
<kbd>s</kbd> - Show/hide pupils<br/>
<kbd>=</kbd> - Increase the number of triangles<br/>
<kbd>-</kbd> - Decrease the number of triangles<br/>
<kbd>]</kbd> - Increase the threshold<br/>
<kbd>[</kbd> - Decrease the threshold<br/>
<kbd>1</kbd> - Increase the stroke size<br/>
<kbd>0</kbd> - Decrease the stroke size<br/>

## Author

* Endre Simo ([@simo_endre](https://twitter.com/simo_endre))

## License

Copyright © 2020 Endre Simo

This software is distributed under the MIT license. See the [LICENSE](https://github.com/esimov/pigo-wasm-demos/blob/master/LICENSE) file for the full license text.
