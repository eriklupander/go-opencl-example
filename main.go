package main

import (
	_ "embed"
	"fmt"
	"github.com/jgillich/go-opencl/cl"
	"unsafe"
)

//go:embed kernel.cl
var kernelSource string

func main() {

	// 1. Set up some input to pass into the OpenCL kernel
	input := make([]int64, 0)
	for i := 0; i < 16; i++ {
		input = append(input, int64(i))
	}
	// size in bytes for each input element. 8, in this case but doing it dynamically looks cooler.
	inputElemSize := int(unsafe.Sizeof(input[0]))

	// 2. Get hold of OpenCL platform and device
	platforms, err := cl.GetPlatforms()
	check("Failed to get platforms", err)

	devices, err := platforms[0].GetDevices(cl.DeviceTypeAll)
	check("Failed to get devices", err)
	if len(devices) == 0 {
		panic("GetDevices returned 0 devices")
	}
	fmt.Println("Using: " + devices[0].Name())
	// 3. Select a device to use. On my mac: 0 == CPU, 1 == Iris GPU, 2 == GeForce 750M GPU
	context, err := cl.CreateContext([]*cl.Device{devices[0]})
	check("CreateContext failed", err)

	// 4. Create a "Command Queue" bound to the first device
	queue, err := context.CreateCommandQueue(devices[0], 0)
	check("CreateCommandQueue failed", err)

	// 5. Create an OpenCL "program" from the source code.
	program, err := context.CreateProgramWithSource([]string{kernelSource})
	check("CreateProgramWithSource failed", err)

	// 5.1 Build the OpenCL program, i.e. compile it.
	err = program.BuildProgram(nil, "")
	check("BuildProgram failed", err)

	// 5.2 Create the actual Kernel with a name, the Kernel is the "function"
	//     we call when we want to execute something.
	kernel, err := program.CreateKernel("square")
	check("CreateKernel failed", err)

	// 6.1 Create OpenCL buffers (memory) for the input. Note that we're allocating 8 bytes per input element, each int64 is 8 bytes in length.
	inputBuffer, err := context.CreateEmptyBuffer(cl.MemReadOnly, inputElemSize*len(input))
	check("CreateBuffer failed for vectors input", err)
	defer inputBuffer.Release()

	// 6.2 Create OpenCL buffer (memory) for the output data, which in our case is identical in length to the input.
	outputBuffer, err := context.CreateEmptyBuffer(cl.MemReadOnly, inputElemSize*len(input))
	check("CreateBuffer failed for output", err)
	defer outputBuffer.Release()

	// 6.3 Connect our input to the command queue and upload the data into device (GPU/CPU) memory. The inputDataPtr is
	// a pointer to the first element of the input slice, while inputDataTotalSizeBytes is the total length of the input data, in bytes
	inputDataPtr := unsafe.Pointer(&input[0])
	inputDataTotalSizeBytes := inputElemSize * len(input)
	_, err = queue.EnqueueWriteBuffer(inputBuffer, true, 0, inputDataTotalSizeBytes, inputDataPtr, nil)
	check("EnqueueWriteBuffer failed", err)

	// 6.4 Kernel is our program and here we explicitly bind our 4 parameters to it
	err = kernel.SetArgs(inputBuffer, outputBuffer)
	check("SetKernelArgs failed", err)

	// 7. Finally, start work! Enqueue executes the loaded args on the specified kernel.
	_, err = queue.EnqueueNDRangeKernel(kernel, nil, []int{len(input)}, []int{16}, nil)
	check("EnqueueNDRangeKernel failed", err)

	// 8. Finish() blocks the main goroutine until the OpenCL queue is empty, i.e. all calculations are done
	err = queue.Finish()
	check("Finish failed", err)

	// 9. Allocate go-side storage for loading the output from the OpenCL program
	results := make([]int64, len(input))

	// 10. EnqueueReadBuffer copies the data in the OpenCL "output" buffer into the "results" slice.
	dataPtrOut := unsafe.Pointer(&results[0])
	sizePerEntry := int(unsafe.Sizeof(results[0]))
	dataSizeOut := sizePerEntry * len(results)

	_, err = queue.EnqueueReadBuffer(outputBuffer, true, 0, dataSizeOut, dataPtrOut, nil)
	check("EnqueueReadBuffer failed", err)

	// 11. We're done! Just dump the results to stdout
	fmt.Printf("%+v\n", results)
}

func check(msg string, err error) {
	if err != nil {
		panic(msg + ": " + err.Error())
	}
}