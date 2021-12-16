# Golang OpenCL example
This is a tiny "hello world"-like application that shows how to use the [github.com/jgillich/go-opencl](github.com/jgillich/go-opencl) OpenCL bindings for Go.

See `main.go` for some relatively well annotated boilerplate code and `kernel.cl` for a simplistic kernel that squares each input number by itself.

Example: (OS X)
```shell
$ go run main.go
Using: Intel(R) Core(TM) i7-4870HQ CPU @ 2.50GHz
[0 1 4 9 16 25 36 49 64 81 100 121 144 169 196 225]
```

### OS X
Seems to work out of the box on the two Macs I've tested the program on.

### Windows 10
CGO_CFLAGS needs to point at the folder containing the `/CL` directory that has the `opencl.h` files.
CGO_LDFLAGS needs to point at the folder containing the graphic card provider's .dll files, such as `nvopencl.dll`.

### Linux
Not tested on Linux