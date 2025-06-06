import os
import json
from datetime import datetime
from collections import defaultdict
import re

def calculate_active_time(entries):
    total_seconds = 0
    start_time = None
    previous_value = None
    for entry in entries:
        value = entry.get('value')
        if value not in (0.0, 1.0):
            continue
        timestamp = datetime.strptime(entry['timestamp'], '%Y-%m-%d %H:%M:%S')
        if previous_value is None:
            previous_value = value
            if value == 1.0:
                start_time = timestamp
            continue
        if previous_value == 0.0 and value == 1.0:
            start_time = timestamp
        elif previous_value == 1.0 and value == 0.0:
            if start_time:
                delta = timestamp - start_time
                total_seconds += delta.total_seconds()
                start_time = None
        previous_value = value
    return total_seconds

def process_all_files(directory, sensor_ids):
    filename_pattern = re.compile(r'(\d+)_(\d{4})-(\d{2})\.json')
    active_time_data = defaultdict(lambda: defaultdict(lambda: defaultdict(int)))
    for filename in os.listdir(directory):
        match = filename_pattern.search(filename)
        if not match:
            continue
        external_id = match.group(1)
        year = int(match.group(2))
        month = int(match.group(3))
        if external_id not in sensor_ids:
            continue
        file_path = os.path.join(directory, filename)
        with open(file_path, 'r') as f:
            entries = json.load(f)
            # Sort entries by timestamp (ascending)
            entries.sort(key=lambda x: datetime.strptime(x['timestamp'], '%Y-%m-%d %H:%M:%S'))
        active_time = calculate_active_time(entries)
        active_time_data[external_id][year][month] = active_time
    return active_time_data

def Than(active_time_data, sensor_ids):
    results = {}
    for sensor in sensor_ids:
        pre = []
        post = []
        for year in active_time_data[sensor]:
            for month in active_time_data[sensor][year]:
                if year < 2025:
                    pre.append(active_time_data[sensor][year][month])
                elif year == 2025 and month >= 1:
                    post.append(active_time_data[sensor][year][month])
        if not pre or not post:
            results[sensor] = {"status": "Insufficient data"}
            continue
        pre_avg = sum(pre) / len(pre) / 3600  # Convert to hours
        post_avg = sum(post) / len(post) / 3600
        increase = ((post_avg - pre_avg) / pre_avg) * 100
        results[sensor] = {
            "pre_avg_hours": round(pre_avg, 2),
            "post_avg_hours": round(post_avg, 2),
            "increase_percent": round(increase, 2)
        }
    return results

def analyze_monthly_trends(active_time_data, target_year=2025):
    pre_year = target_year - 1
    results = {}

    for sensor_id, years_data in active_time_data.items():
        monthly_changes = []

        for month in range(1, 13):  # For all 12 months
            pre_hours = years_data.get(pre_year, {}).get(month)
            post_hours = years_data.get(target_year, {}).get(month)

            if pre_hours is not None and post_hours is not None:
                if pre_hours > 0:  # Avoid division by zero
                    change_pct = ((post_hours - pre_hours) / pre_hours) * 100
                    monthly_changes.append((month, pre_hours, post_hours, change_pct))

        if monthly_changes:
            total_pre = sum(p for _, p, _, _ in monthly_changes)
            total_post = sum(po for _, _, po, _ in monthly_changes)
            overall_change_pct = ((total_post - total_pre) / total_pre) * 100

            results[sensor_id] = {
                "monthly_details": monthly_changes,
                "pre_avg_hours": total_pre / len(monthly_changes),
                "post_avg_hours": total_post / len(monthly_changes),
                "overall_change_pct": round(overall_change_pct, 2)
            }
        else:
            results[sensor_id] = {"status": "Insufficient data for comparison"}

    return results

def main():
    # 1. Define the directory containing your JSON files
    data_directory = "trenddata"  # Update this path

    # 2. List of sensor IDs to process
    sensor_ids = ["176558", "176567"]  # Update with your actual sensor IDs

    # 3. Process all files and calculate active time
    active_time_data = process_all_files(data_directory, sensor_ids)

    # 4. Calculate statistics
    # results = Than(active_time_data, sensor_ids)

    # 5. Output results
    #print(json.dumps(results, indent=2))

    results = analyze_monthly_trends(active_time_data, target_year=2025)
    for sensor_id, data in results.items():
        print(f"Sensor {sensor_id}:")
        if "status" in data:
            print(f"  {data['status']}")
        else:
            print(f"  Pre-Avg (2024): {data['pre_avg_hours']:.2f}h")
            print(f"  Post-Avg (2025): {data['post_avg_hours']:.2f}h")
            print(f"  Overall Change: {data['overall_change_pct']:.2f}%")
            print("  Monthly Changes:")
            for month, pre, post, pct in data["monthly_details"]:
                print(f"    Month {month}: {pre:.2f}h → {post:.2f}h ({pct:.2f}%)")

if __name__ == "__main__":
    main()
