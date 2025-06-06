
import os
import json
import csv
from datetime import datetime, timedelta

# Set this to the folder where your JSON files are located
data_folder = "trenddata"
output_csv = "active_hours_summary.csv"

def parse_timestamp(entry):
    try:
        return datetime.strptime(entry["timestamp"], "%Y-%m-%d %H:%M:%S")
    except Exception:
        return None

def process_file(filepath):
    with open(filepath, "r") as f:
        data = json.load(f)

    filtered = [entry for entry in data if entry["value"] in [0.0, 1.0]]
    filtered.sort(key=lambda x: parse_timestamp(x))

    active_periods = []
    active_start = None

    for entry in filtered:
        timestamp = parse_timestamp(entry)
        if timestamp is None:
            continue

        if entry["value"] == 1.0 and active_start is None:
            active_start = timestamp
        elif entry["value"] == 0.0 and active_start is not None:
            active_periods.append((active_start, timestamp))
            active_start = None

    total_active_time = sum((end - start for start, end in active_periods), timedelta())
    return total_active_time.total_seconds() / 3600

def write_results_to_csv(results, output_path):
    with open(output_path, mode="w", newline="") as csvfile:
        writer = csv.writer(csvfile)
        writer.writerow(["Filename", "Active Hours"])
        for filename, hours in sorted(results.items()):
            writer.writerow([filename, round(hours, 2)])

def main():
    if not os.path.exists(data_folder):
        print(f"Folder not found: {data_folder}")
        return

    results = {}
    for filename in os.listdir(data_folder):
        if filename.endswith(".json"):
            filepath = os.path.join(data_folder, filename)
            hours = process_file(filepath)
            results[filename] = hours

    write_results_to_csv(results, os.path.join(data_folder, output_csv))
    print(f"Results written to {output_csv}")

if __name__ == "__main__":
    main()
