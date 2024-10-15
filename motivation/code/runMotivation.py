import subprocess
import time
import os
from multiprocessing import Process
from sendRequests import send_requests_to_all_pods
from getResult import log_processing_results
from kubernetes import client, config
from datetime import datetime

# YAML 파일의 경로와 네임스페이스 정의
NAMESPACE = "pair"
DOWNSTREAM_YAML_PATH = "/home/dnc/master/Benchmarks/pair/deployments/downstream.yaml"
DOWNSTREAM_SERVICE = "downstream"
UPSTREAM_SERVICE = "upstream"
QPS = 5  # 초당 10개의 요청
DURATION = 60  # 60초 동안 반복
RESULT_PATH = "/home/dnc/master/CCgrid2024/motivation/results"  # 결과 저장 경로

# `kubectl` 명령어 실행을 위한 함수
def kubectl_delete_and_apply(yaml_path):
    try:
        # 기존 리소스 삭제
        print(f"Deleting resources from {yaml_path}...")
        subprocess.run(f"kubectl delete -Rf {yaml_path} --force", shell=True, check=True)
        time.sleep(30)  # 30초 대기
        # 리소스 다시 생성
        print(f"Applying resources from {yaml_path}...")
        subprocess.run(f"kubectl apply -f {yaml_path}", shell=True, check=True)
        
        # 파드가 준비될 때까지 대기
        print("Waiting for 30 seconds to let the pods become ready...")
        time.sleep(30)  # 30초 대기
    except subprocess.CalledProcessError as e:
        print(f"Error executing kubectl command: {e}")
        exit(1)

# Kubernetes 클라이언트 설정
def get_k8s_endpoints(service_name, namespace):
    config.load_kube_config()  # kubeconfig 로드
    v1 = client.CoreV1Api()
    endpoints = v1.read_namespaced_endpoints(service_name, namespace)
    
    ip_list = []
    for subset in endpoints.subsets:
        for address in subset.addresses:
            ip_list.append(f"{address.ip}:{subset.ports[0].port}")
    return ip_list

# Kubernetes 클라이언트 설정 및 Metrics API를 통한 CPU 사용량 수집
def get_pod_cpu_usage(namespace):
    config.load_kube_config()  # kubeconfig 로드
    metrics_api = client.CustomObjectsApi()
    metrics = metrics_api.list_namespaced_custom_object(
        group="metrics.k8s.io",
        version="v1beta1",
        namespace=namespace,
        plural="pods"
    )

    cpu_usage_data = {}
    for pod in metrics['items']:
        pod_name = pod['metadata']['name']
        cpu_usage = pod['containers'][0]['usage']['cpu']
        cpu_millicores = int(cpu_usage[:-1]) if cpu_usage.endswith('n') else int(cpu_usage[:-1]) * 1000
        cpu_millicores = cpu_millicores // 1_000_000  # 소수점 이하 6자리 버림
        cpu_usage_data[pod_name] = cpu_millicores
    return cpu_usage_data

# 요청을 보내는 함수
def send_requests(upstream_ips):
    start_time = time.time()
    elapsed_time = 0

    print("Starting to send requests to upstream pods for 60 seconds...")
    while elapsed_time < DURATION:
        send_requests_to_all_pods(upstream_ips, QPS)  # 초당 10개의 요청 전송
        time.sleep(1)  # 1초 대기
        elapsed_time = time.time() - start_time

# CPU 사용량을 10초마다 수집하여 기록하는 함수
def monitor_cpu_usage(namespace, result_file):
    total_cpu_usage = {}
    num_samples = 0

    start_time = time.time()
    with open(result_file, 'w') as f:
        while time.time() - start_time < DURATION + 60:  # 30초 먼저 모니터링 시작
            # 매 10초마다 CPU 사용량 수집
            current_cpu_usage = get_pod_cpu_usage(namespace)
            f.write(f"--- 10-second CPU usage at {datetime.now().strftime('%H:%M:%S')} ---\n")
            for pod_name, cpu_usage in current_cpu_usage.items():
                if pod_name not in total_cpu_usage:
                    total_cpu_usage[pod_name] = 0
                total_cpu_usage[pod_name] += cpu_usage
                f.write(f"Pod: {pod_name}, CPU Usage: {cpu_usage} millicores\n")
            num_samples += 1
            f.write("\n")
            time.sleep(10)  # 10초 대기

        # 1분간 평균 CPU 사용량 계산
        if num_samples > 0:
            f.write("--- Average CPU usage over 1 minute ---\n")
            for pod_name, total_usage in total_cpu_usage.items():
                avg_usage = total_usage // num_samples
                f.write(f"Pod: {pod_name}, Avg CPU Usage: {avg_usage} millicores\n")
            f.write("\n")

if __name__ == "__main__":
    # YAML 파일을 삭제하고 다시 적용
    kubectl_delete_and_apply(DOWNSTREAM_YAML_PATH)

    # 현재 시간으로 파일 저장 경로 생성
    current_time = datetime.now().strftime("%Y%m%d_%H%M%S")
    result_file = os.path.join(RESULT_PATH, f"results_{current_time}.txt")
    
    # CPU 사용량 모니터링 프로세스 시작
    monitor_process = Process(target=monitor_cpu_usage, args=(NAMESPACE, result_file))
    monitor_process.start()
    
    # 모니터링을 30초 동안 수행 후 요청 프로세스 시작
    time.sleep(30)
    
    # Kubernetes에서 Downstream과 Upstream IP 가져오기
    downstream_ips = get_k8s_endpoints(DOWNSTREAM_SERVICE, NAMESPACE)
    upstream_ips = get_k8s_endpoints(UPSTREAM_SERVICE, NAMESPACE)
    
    # 요청을 보내는 프로세스 실행
    request_process = Process(target=send_requests, args=(upstream_ips,))
    request_process.start()

    # 두 프로세스가 완료될 때까지 대기
    request_process.join()
    monitor_process.join()

    # Downstream 처리 결과 조회 및 저장
    print("Retrieving logs from downstream pods...")
    downstream_stats = log_processing_results(downstream_ips)

    time.sleep(30)

    # Downstream 로그 및 CPU 사용량 결과 파일에 저장
    with open(result_file, 'a') as f:
        for downstream_ip, stats in downstream_stats.items():
            f.write(f"Stats from {downstream_ip}:\n")
            f.write(f"Average Process Time: {stats['average_time']}ms\n")
            f.write(f"Total Received Requests: {stats['total_received']}\n")
            
            # 각 로그 (요청 번호 및 처리 시간) 출력
            for log in stats['logs']:
                f.write(f"{log}\n")
            f.write("\n")

    print(f"Results and CPU usage saved to {result_file}")
