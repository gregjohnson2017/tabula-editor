# tabula-editor
Greg and Tyler's editor

### Dependencies:
#### Linux:
```
sudo apt install libsdl2-ttf-dev libsdl2-gfx-dev
```

#### Windows:

Install gcc for C Go using msys2.

After installing msys2, start the msys2 shell (not MinGW) and run
```
pacman -Syu
```
and then
```
pacman -S mingw-w64-x86_64-SDL2 mingw-w64-x86_64-SDL2_gfx mingw-w64-x86_64-SDL2_ttf mingw-w64-x86_64-gcc
```
After everything is finished, add the mingw bin to your PATH, e.g., C:\msys64\mingw64\bin.
