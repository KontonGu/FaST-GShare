#include <iostream>
#include <cuda_runtime.h>

#define N 512  // 矩阵的维度

__global__ void matrixAdd(float* A, float* B, float* C, int width) {
    int x = blockIdx.x * blockDim.x + threadIdx.x;
    int y = blockIdx.y * blockDim.y + threadIdx.y;

    if (x < width && y < width) {
        int index = y * width + x;
        C[index] = A[index] + B[index];
    }
}

int main() {
    // 矩阵大小
    int size = N * N * sizeof(float);
    
    // 主机端数据
    float *h_A, *h_B, *h_C;

    // 设备端数据
    float *d_A, *d_B, *d_C;

    // 分配主机内存
    h_A = (float*)malloc(size);
    h_B = (float*)malloc(size);
    h_C = (float*)malloc(size);

    // 初始化矩阵 A 和 B
    for (int i = 0; i < N * N; i++) {
        h_A[i] = static_cast<float>(i);
        h_B[i] = static_cast<float>(i);
    }

    // 分配设备内存
    cudaMalloc((void**)&d_A, size);
    cudaMalloc((void**)&d_B, size);
    cudaMalloc((void**)&d_C, size);

    // 将数据从主机复制到设备
    cudaMemcpy(d_A, h_A, size, cudaMemcpyHostToDevice);
    cudaMemcpy(d_B, h_B, size, cudaMemcpyHostToDevice);

    // 设定线程和块的数量
    dim3 threadsPerBlock(16, 16);
    dim3 numBlocks((N + threadsPerBlock.x - 1) / threadsPerBlock.x, 
                   (N + threadsPerBlock.y - 1) / threadsPerBlock.y);

    // 调用内核函数进行矩阵加法
    matrixAdd<<<numBlocks, threadsPerBlock>>>(d_A, d_B, d_C, N);

    // 释放不必要的 GPU 内存（例如不再需要的 d_A 和 d_B）
    cudaFree(d_A);
    cudaFree(d_B);

    // 从设备复制结果到主机
    cudaMemcpy(h_C, d_C, size, cudaMemcpyDeviceToHost);

    // 打印前几个结果以验证
    for (int i = 0; i < 10; i++) {
        std::cout << h_C[i] << " ";
    }
    std::cout << std::endl;

    // 释放设备内存
    cudaFree(d_C);

    // 释放主机内存
    free(h_A);
    free(h_B);
    free(h_C);

    return 0;
}