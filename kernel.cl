__kernel void square(__global long *input, __global long *output) {
    int i = get_global_id(0);
    output[i] = input[i]*input[i];
}