import requests

# 각 파드별 처리 결과 반환 (log 대신 stats로 요청)
def log_processing_results(downstream_ips):
    results = {}
    
    for downstream_ip in downstream_ips:
        stats_url = f"http://{downstream_ip}/stats"  # /log 대신 /stats로 변경
        try:
            response = requests.get(stats_url)
            if response.status_code == 200:
                stats = response.json()

                # 평균 처리 시간 및 총 요청 수 가져오기
                average_time = stats.get('average_process_time_ms', 'N/A')
                total_received = stats.get('total_received_requests', 'N/A')
                logs = stats.get('logs', [])

                # 결과를 딕셔너리에 저장
                results[downstream_ip] = {
                    'average_time': average_time,
                    'total_received': total_received,
                    'logs': logs
                }
                
                print(f"Stats from {downstream_ip} retrieved successfully.")
            else:
                print(f"Failed to retrieve stats from {downstream_ip}: {response.status_code}")
        except requests.exceptions.RequestException as e:
            print(f"Error retrieving stats from {downstream_ip}: {e}")
    
    return results
