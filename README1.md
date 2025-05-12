## HAS-GPU: Efficient Hybrid Auto-scaling with Fine-grained GPU Allocation for SLO-aware Serverless Inferences

### Hardware and System Setup
- LRZ Compute Cloud VMs [(website)](https://doku.lrz.de/compute-cloud-10333232.html);
- NVIDIA Tesla V100 16 GB RAM; (10 VM nodes, each with 1 GPU);
- Ubuntu 20.04; (Only tested)
- NVIDIA Driver 535.54.03 , CUDA 12.2;
- Helm v3.17.0
- Kubernetes 1.32.1;

---
### Part-1: HAS-GPU Deployment
#### 1. Prepare K8s infrastructure environment;

#### 2. Deploy FaSTPod system that HAS-GPU System relys on;

#### 3. Deploy HAS-GPU

#### 4. Run Experiment Workload


---
### Part-2: Other systems for experimental comparison: KServe, FaST-GShare. (Optional)

#### 1. KServe (Deploy based on new K8s environment)


#### 2. FaST-GShare


---
### Part-3: Build HAS-GPU from Scratch (Optional)
