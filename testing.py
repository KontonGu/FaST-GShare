#!/usr/bin/env python3
import subprocess
import time
import os
import sys
import datetime
import argparse


def run_command(command, error_message=None):
    """Run a shell command and handle errors"""
    print(f"Executing: {command}")
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Error: {error_message or 'Command failed'}")
        print(f"Error output: {result.stderr}")
        sys.exit(1)
    return result.stdout.strip()


def wait_for_pod_ready(pod_pattern, namespace="default", timeout=300):
    """Wait for a pod matching the pattern to be in Ready state"""
    print(f"Waiting for pod matching '{pod_pattern}' to be ready...")
    
    start_time = time.time()
    while time.time() - start_time < timeout:
        result = subprocess.run(
            f"kubectl get pods -n {namespace} | grep {pod_pattern}", 
            shell=True, 
            capture_output=True, 
            text=True
        )
        
        if result.returncode == 0:
            # Pod exists, check if it's ready
            pod_info = result.stdout.strip().split()
            if len(pod_info) >= 2:
                status = pod_info[1]  # The second column is typically the STATUS
                if "Running" in status and "1/1" in status:
                    print(f"Pod '{pod_pattern}' is ready!")
                    return True
                else:
                    print(f"Pod status: {status}, waiting...")
        
        time.sleep(5)
    
    print(f"Timed out waiting for pod '{pod_pattern}' to be ready")
    sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description="Run HAS deployment and testing")
    parser.add_argument("--output-filename", required=True, help="Specify the K6 CSV output filename")
    parser.add_argument("--copy-csv-suffix", required=True, help="Suffix for the HAS usage CSV file")
    args = parser.parse_args()
    
    # Create a timestamp
    timestamp = datetime.datetime.now().strftime("%Y%m%d-%H%M%S")
    
    # 1. Execute make has_deploy and wait for pod1 to be ready
    print("Step 1: Deploying HAS service...")
    run_command("make has_deploy", "Failed to deploy HAS service")
    wait_for_pod_ready("has-deployment")
    
    # 2. Execute make has_sample and wait for pod2 to be ready
    print("Step 2: Deploying HAS sample...")
    run_command("make has_sample", "Failed to deploy HAS sample")
    wait_for_pod_ready("has-sample")
    
    # 3. Run k6 test with specified output filename
    print(f"Step 3: Running k6 test with output to {args.output_filename}...")
    k6_cmd = f"k6 run --out csv={args.output_filename} k6-trace.js > k6.txt"
    run_command(k6_cmd, "Failed to run k6 test")
    
    # 4. Remove HAS sample
    print("Step 4: Removing HAS sample...")
    run_command("make has_sample_remove", "Failed to remove HAS sample")
    
    # 5. Copy CSV file
    csv_copy_cmd = f"cp /data/dir/has-usage.csv /dir/copy-to/has-usage-{args.copy_csv_suffix}.csv"
    print(f"Step 5: Copying CSV file: {csv_copy_cmd}")
    run_command(csv_copy_cmd, "Failed to copy CSV file")
    
    # 6. Undeploy HAS
    print("Step 6: Undeploying HAS...")
    run_command("make has_undeploy", "Failed to undeploy HAS")
    
    print("All steps completed successfully!")


if __name__ == "__main__":
    main()