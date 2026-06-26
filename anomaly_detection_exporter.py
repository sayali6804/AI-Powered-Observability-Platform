
# import requests
# import pandas as pd
# from sklearn.ensemble import IsolationForest
# from sklearn.preprocessing import StandardScaler
# import time
# from prometheus_client import start_http_server, Gauge

# # Prometheus query to fetch metric data — replace with your metric name
# PROMETHEUS_QUERY = 'up'  # example metric

# # Create Prometheus Gauge metric with label metric_name only (no timestamp label)
# anomaly_gauge = Gauge('anomaly_detected', 'Anomaly detected flag (1=anomaly)', ['metric_name'])

# def clean_metric_name(name):
#     # Remove problematic chars like colon and dots, replace by underscore
#     return name.replace(':', '_').replace('.', '_')

# def fetch_metrics(query, duration_seconds=3600, step_seconds=60):
#     """
#     Fetch metric data from Prometheus over the last `duration_seconds`
#     """
#     end = int(time.time())
#     start = end - duration_seconds

#     url = 'http://localhost:9090/api/v1/query_range'
#     params = {
#         'query': query,
#         'start': start,
#         'end': end,
#         'step': step_seconds
#     }

#     resp = requests.get(url, params=params)
#     resp.raise_for_status()
#     data = resp.json()['data']['result']

#     # Build DataFrame with timestamps as index and metric series as columns
#     df = pd.DataFrame()
#     for series in data:
#         metric_labels = "_".join([f"{k}_{v}" for k, v in series['metric'].items()])
#         values = series['values']
#         ts_df = pd.DataFrame(values, columns=['timestamp', metric_labels])
#         ts_df['timestamp'] = pd.to_datetime(ts_df['timestamp'], unit='s')
#         ts_df.set_index('timestamp', inplace=True)
#         df = pd.concat([df, ts_df], axis=1)

#     # Convert all columns to numeric, forward/backward fill missing values
#     df = df.apply(pd.to_numeric, errors='coerce')
#     df = df.ffill().bfill()

#     return df

# def detect_anomalies(df):
#     scaler = StandardScaler()
#     X_scaled = scaler.fit_transform(df)

#     model = IsolationForest(contamination=0.05, random_state=42)
#     model.fit(X_scaled)

#     df['anomaly'] = model.predict(X_scaled)
#     # IsolationForest labels anomalies as -1, normal as 1
#     df['anomaly'] = df['anomaly'].apply(lambda x: 1 if x == -1 else 0)
#     return df

# def export_anomalies(df):
#     """
#     Export anomalies to Prometheus Gauge metric
#     """
#     # Clear previous metric values to avoid stale data
#     anomaly_gauge.clear()

#     for ts, row in df.iterrows():
#         anomaly_flag = row['anomaly']
#         for col in df.columns:
#             if col != 'anomaly':
#                 metric_label = clean_metric_name(col)
#                 anomaly_gauge.labels(metric_name=metric_label).set(anomaly_flag)

# def main():
#     # Start Prometheus HTTP server on port 8000
#     start_http_server(8000)
#     print("[*] Prometheus metrics available at http://localhost:8000/metrics")

#     while True:
#         print("[*] Fetching Prometheus metrics...")
#         df = fetch_metrics(PROMETHEUS_QUERY)

#         if df.empty:
#             print("[!] No data found for the metric. Check your query and Prometheus setup.")
#             time.sleep(60)
#             continue

#         print("[*] Running anomaly detection...")
#         df = detect_anomalies(df)

#         print(df.tail())
#         df.to_csv("anomaly_results.csv")
#         print("[+] Anomaly detection results saved to 'anomaly_results.csv'")

#         # Export anomaly data to Prometheus metrics endpoint
#         export_anomalies(df)
#         print(f"[+] Exported anomaly metrics at {time.strftime('%Y-%m-%d %H:%M:%S')}")

#         time.sleep(60)  # run every 60 seconds

# if __name__ == '__main__':
#     main()


import requests
import pandas as pd
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler
import time
from prometheus_client import start_http_server, Gauge

# Prometheus query to fetch metric data — replace with your metric name
PROMETHEUS_QUERY = 'up'  # example metric

# Create Prometheus Gauge metric with label metric_name only (no timestamp label)
anomaly_gauge = Gauge('anomaly_detected', 'Anomaly detected flag (1=anomaly)', ['metric_name'])

def clean_metric_name(name):
    # Remove problematic chars like colon and dots, replace by underscore
    return name.replace(':', '_').replace('.', '_').replace(' ', '_')

def fetch_metrics(query, duration_seconds=3600, step_seconds=60):
    """
    Fetch metric data from Prometheus over the last `duration_seconds`
    """
    end = int(time.time())
    start = end - duration_seconds

    url = 'http://localhost:9090/api/v1/query_range'
    params = {
        'query': query,
        'start': start,
        'end': end,
        'step': step_seconds
    }

    resp = requests.get(url, params=params)
    resp.raise_for_status()
    data = resp.json()['data']['result']

    # Build DataFrame with timestamps as index and metric series as columns
    df = pd.DataFrame()
    for series in data:
        metric_labels = "_".join([f"{k}_{v}" for k, v in series['metric'].items()])
        values = series['values']
        ts_df = pd.DataFrame(values, columns=['timestamp', metric_labels])
        ts_df['timestamp'] = pd.to_datetime(ts_df['timestamp'], unit='s')
        ts_df.set_index('timestamp', inplace=True)
        df = pd.concat([df, ts_df], axis=1)

    # Convert all columns to numeric, forward/backward fill missing values
    df = df.apply(pd.to_numeric, errors='coerce')
    df = df.ffill().bfill()

    return df

def detect_anomalies(df):
    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(df)

    model = IsolationForest(contamination=0.05, random_state=42)
    model.fit(X_scaled)

    df['anomaly'] = model.predict(X_scaled)
    # IsolationForest labels anomalies as -1, normal as 1
    df['anomaly'] = df['anomaly'].apply(lambda x: 1 if x == -1 else 0)
    return df

def export_anomalies(df):
    """
    Export anomalies to Prometheus Gauge metric
    """
    anomaly_gauge.clear()
    for ts, row in df.iterrows():
        anomaly_flag = row['anomaly']
        for col in df.columns:
            if col != 'anomaly':
                metric_label = clean_metric_name(col)
                anomaly_gauge.labels(metric_name=metric_label).set(anomaly_flag)
                print(f"Exported anomaly_detected metric_name={metric_label} value={anomaly_flag}")

def main():
    # Start Prometheus HTTP server on port 8000
    start_http_server(8000)
    print("[*] Prometheus metrics available at http://localhost:8000/metrics")

    while True:
        print("[*] Fetching Prometheus metrics...")
        df = fetch_metrics(PROMETHEUS_QUERY)

        if df.empty:
            print("[!] No data found for the metric. Check your query and Prometheus setup.")
            time.sleep(60)
            continue

        print("[*] Running anomaly detection...")
        df = detect_anomalies(df)

        # Debug print for anomaly counts
        print("[*] Anomaly value counts:", df['anomaly'].value_counts())

        # Optional: force anomaly on first row for testing (comment this out if not needed)
        df.loc[df.index[0], 'anomaly'] = 1

        print(df.tail())
        df.to_csv("anomaly_results.csv")
        print("[+] Anomaly detection results saved to 'anomaly_results.csv'")

        export_anomalies(df)
        print(f"[+] Exported anomaly metrics at {time.strftime('%Y-%m-%d %H:%M:%S')}")

        time.sleep(60)  # run every 60 seconds

if __name__ == '__main__':
    main()
