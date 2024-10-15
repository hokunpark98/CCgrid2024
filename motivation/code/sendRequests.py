import requests
import threading


def send_request_to_pod(pod_ip, QPS):
    url = f"http://{pod_ip}/send?value={QPS}"
    try:
        response = requests.post(url)
        if response.status_code == 200:
            print(f"Successfully sent request to {pod_ip}: {response.text}")
        else:
            print(f"Failed to send request to {pod_ip}: {response.status_code}")
    except requests.exceptions.RequestException as e:
        print(f"Error sending request to {pod_ip}: {e}")

# 모든 파드에 동시에 요청 보내기
def send_requests_to_all_pods(pod_ips, QPS):
    threads = []
    for pod_ip in pod_ips:
        thread = threading.Thread(target=send_request_to_pod, args=(pod_ip, QPS))
        threads.append(thread)
        thread.start()

    # 모든 스레드가 완료될 때까지 대기
    for thread in threads:
        thread.join()
