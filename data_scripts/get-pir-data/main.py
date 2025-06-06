
import csv
import json
from collections import defaultdict
from datetime import datetime

# === CONFIGURATION ===
csv_file = "active_hours_summary.csv"
room_mapping_json = "getRoomActivity.json"
start_month = "2025-01"  # <-- Change this to any YYYY-MM string
# =====================


def load_room_mapping(json_file):
    """Builds mapping from externalID (as string) to roomNumber."""
    with open(json_file, "r", encoding="utf-8") as f:
        data = json.load(f)

    mapping = {}
    for building in data.get("data", {}).get("buildings", []):
        for floor in building.get("floors", []):
            for room in floor.get("rooms", []):
                for sensor in room.get("sensors", []):
                    ext_id = str(sensor.get("externalID"))
                    room_number = room.get("roomNumber")
                    if ext_id and room_number:
                        mapping[ext_id] = room_number
    return mapping


def extract_info_from_filename(filename):
    """Parses filename like 175882_2022-10.json into externalID and year-month."""
    if not filename.endswith(".json"):
        return None, None
    try:
        base = filename.replace(".json", "")
        ext_id, ym = base.split("_")
        datetime.strptime(ym, "%Y-%m")  # validate date format
        return ext_id, ym
    except Exception:
        return None, None


def load_activity_data(csv_file, room_map, start_month_str):
    """Aggregates hours per room from CSV based on start month."""
    start_month = datetime.strptime(start_month_str, "%Y-%m")
    total_hours = defaultdict(float)

    with open(csv_file, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            filename = row["Filename"]
            hours = float(row["Active Hours"])

            ext_id, ym = extract_info_from_filename(filename)
            if ext_id is None or ym is None:
                continue

            row_date = datetime.strptime(ym, "%Y-%m")
            if row_date >= start_month and ext_id in room_map:
                room = room_map[ext_id]
                total_hours[room] += hours

    return total_hours


def print_ranked_usage(hours_by_room):
    sorted_usage = sorted(hours_by_room.items(), key=lambda x: x[1], reverse=True)

    print("\n🔝 Top 10 Most Active Rooms:")
    for room, hours in sorted_usage[:10]:
        print(f"{room}: {hours:.1f} hours")

    print("\n🔻 Top 5 Least Active Rooms:")
    for room, hours in sorted_usage[-5:]:
        print(f"{room}: {hours:.1f} hours")


if __name__ == "__main__":
    print(f"🔍 Analyzing room activity from {start_month} onward...")

    room_map = load_room_mapping(room_mapping_json)
    hours_by_room = load_activity_data(csv_file, room_map, start_month)
    print_ranked_usage(hours_by_room)
