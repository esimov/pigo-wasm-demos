# pigo-wasm-demos

<p align="center"><img src="https://user-images.githubusercontent.com/883386/80915158-06911a80-8d59-11ea-93bd-eca98750ad62.png" alt="Pigo Wasm demos" title="Pigo Wasm demos" width="400"/></p>

This repository is a collection of Webassembly demos showcasing a few examples of the [Pigo](https://github.com/esimov/pigo) face detection library running real time. **This repo will be continuously updated**.

## Install
**Notice: at least Go 1.13 is required in order to run the demos!**

```bash
$ go install github.com/esimov/pigo-wasm-demos@latest
```

## Run

You only need to type `$make demo{no}`. This will build the package and produce an executable WebAssembly file which can be served over an http server. A new tab will be opened automatically in the user's default browser. 

## Demos

### Masquerade
```bash
$ make demo1
```
![pigo_wasm_masquarade](https://user-images.githubusercontent.com/883386/82048111-ae450b80-96bc-11ea-9f22-7039ce937140.gif)


#### Key bindings:
<kbd>q</kbd> - Show/hide the detected face rectangle<br/>
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
![pigo_wasm_faceblur](https://user-images.githubusercontent.com/883386/170483688-5a145550-5a7b-4400-af34-842333fb1a8e.gif)

#### Key bindings:
<kbd>]</kbd> - Increase the blur radius<br/>
<kbd>[</kbd> - Decrease the blur radius<br/>
<kbd>f</kbd> - Show/hide the detected face rectangle<br/>
<kbd>s</kbd> - Show/hide pupils<br/>
<kbd>b</kbd> - Enable/disable face blur<br/>

### Background blur (in Zoom style)
```bash
$ make demo3
```
![pigo_wasm_background_blur](https://user-images.githubusercontent.com/883386/170483670-2ad0f865-d89d-44c4-8cb5-f9d5736d12fb.gif)

#### Key bindings:
<kbd>]</kbd> - Increase the blur radius<br/>
<kbd>[</kbd> - Decrease the blur radius<br/>
<kbd>f</kbd> - Show/hide the detected face rectangle<br/>
<kbd>s</kbd> - Show/hide pupils<br/>

### Face triangulator
```bash
$ make demo4
```
![pigo_wasm_triangulate](https://user-images.githubusercontent.com/883386/170484192-c43bafa5-36c6-41a8-9e23-3f3d04264b08.gif)

#### Key bindings:
<kbd>f</kbd> - Show/hide the detected face rectangle<br/>
<kbd>s</kbd> - Show/hide pupils<br/>
<kbd>=</kbd> - Increase the number of triangles<br/>
<kbd>-</kbd> - Decrease the number of triangles<br/>
<kbd>]</kbd> - Increase the threshold<br/>
<kbd>[</kbd> - Decrease the threshold<br/>
<kbd>1</kbd> - Increase the stroke size<br/>
<kbd>0</kbd> - Decrease the stroke size<br/>


### Pixelate
```bash
$ make demo5
```
![pigo_wasm_pixelate](https://user-images.githubusercontent.com/883386/170484527-b98745e5-5f93-45cb-a86d-ed12332c8d41.gif)

#### Key bindings:
<kbd>f</kbd> - Show/hide the detected face rectangle<br/>
<kbd>s</kbd> - Show/hide pupils<br/>
<kbd>=</kbd> - Increase the number of colors<br/>
<kbd>-</kbd> - Decrease the number of colors<br/>
<kbd>]</kbd> - Increase the cells size<br/>
<kbd>[</kbd> - Decrease the cells size<br/>

### Triangulated facemask
```bash
$ make demo6
```
![facemask](https://user-images.githubusercontent.com/883386/170938798-9bc7b9b1-ffd4-4add-a536-057c11542991.gif)

This demo is meant to be a proof of concept for an idea of generating personalized triangulated face masks. The rectangle at the top right corner of the screen will turn green when the head alignment is the most appropriate for making a screen capture and this is when the head is aligned perpendicular (+/- a predefined threshold) and close enough to the camera. This demo can be expanded way further.

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
