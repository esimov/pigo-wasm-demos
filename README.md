# pigo-wasm-demos

<p align="center"><img src="https://user-images.githubusercontent.com/883386/80915158-06911a80-8d59-11ea-93bd-eca98750ad62.png" alt="Pigo Wasm demos" title="Pigo Wasm demos" width="400"/></p>

This repository is a collection of real time Webassembly demos based on the [Pigo](https://github.com/esimov/pigo) face detection library, showing various use cases the library can be used for. It's continuously updated.

## Install
**Notice: at least Go 1.13 is required prior running the demos!**

```bash
$ go get -u -v github.com/esimov/pigo-wasm-demos 

```

## Run
To run the demos is as simple as typing a single line of command. This will build the package and produce an executable WebAssembly file which can be served over a http server. A new window will be opened in the user's default browser.

```bash
$ make demo<number>
```

Check the `Makefile` for the existing demos.

## Author

* Endre Simo ([@simo_endre](https://twitter.com/simo_endre))

## License

Copyright © 2020 Endre Simo

This software is distributed under the MIT license. See the [LICENSE](https://github.com/esimov/pigo-wasm-demos/blob/master/LICENSE) file for the full license text.
