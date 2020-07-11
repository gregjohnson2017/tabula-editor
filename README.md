# tabula-editor
Greg and Tyler's editor

### Installation:

You can clone it and build it or go get it.

You can skip many of the dependencies for SDL by cloning and adding the `-tags static` build flag, e.g.:
```
go build -v -tags static -ldflags "-s -w"
```

Non-static builds require the dependencies listed below.

### Dependencies:
#### Linux:
```
sudo apt install libsdl2-dev
```

#### Windows:

Install msys2.

Then, start the msys2 shell (not MinGW) and run
```
pacman -Syu
```
to get up to date and then
```
pacman -S mingw-w64-x86_64-SDL2 mingw-w64-x86_64-gcc
```
to install gcc and SDL2.

After everything is finished, add the mingw bin to your PATH, e.g., `C:\msys64\mingw64\bin`.
