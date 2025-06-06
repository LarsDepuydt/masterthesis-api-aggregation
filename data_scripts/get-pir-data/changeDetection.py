import csv
import json
from collections import defaultdict
from datetime import datetime
from typing import List, Dict


# === CONFIGURATION ===
csv_file = "active_hours_summary.csv"
room_mapping_json = "getRoomActivity.json"
start_month = "2023-01"  # Change this as needed
rooms_to_track = ["A.101a", "A.101b"]  # Add any roomNumbers here
# =====================


def load_room_to_external_ids(json_file: str) -> Dict[str, List[str]]:
    """Returns mapping from roomNumber to list of externalIDs."""
    with open(json_file, "r", encoding="utf-8") as f:
        data = json.load(f)

    mapping = defaultdict(list)
    for building in data.get("data", {}).get("buildings", []):
        for floor in building.get("floors", []):
            for room in floor.get("rooms", []):
                room_number = room.get("roomNumber")
                for sensor in room.get("sensors", []):
                    ext_id = str(sensor.get("externalID"))
                    if room_number and ext_id:
                        mapping[room_number].append(ext_id)
    return mapping


def extract_info_from_filename(filename: str):
    """Extract externalID and month string (YYYY-MM) from filename."""
    if not filename.endswith(".json"):
        return None, None
    try:
        base = filename.replace(".json", "")
        ext_id, ym = base.split("_")
        datetime.strptime(ym, "%Y-%m")  # validate date
        return ext_id, ym
    except Exception:
        return None, None


def build_monthly_usage(csv_file: str, room_to_ids: Dict[str, List[str]], start_month_str: str, rooms: List[str]):
    start_date = datetime.strptime(start_month_str, "%Y-%m")
    usage = defaultdict(lambda: defaultdict(float))  # room -> month -> hours

    tracked_ids = {rid for room in rooms if room in room_to_ids for rid in room_to_ids[room]}

    with open(csv_file, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            filename = row["Filename"]
            hours = float(row["Active Hours"])
            ext_id, month = extract_info_from_filename(filename)

            if ext_id and month and ext_id in tracked_ids:
                if datetime.strptime(month, "%Y-%m") >= start_date:
                    for room, ids in room_to_ids.items():
                        if ext_id in ids:
                            usage[room][month] += hours
    return usage


def print_monthly_usage_with_diff(usage: Dict[str, Dict[str, float]], rooms_to_track: List[str]):
    print("\n📊 Monthly Room Usage (vs. same month previous year)")

    # Get all unique months from the usage data and sort them
    all_months = sorted(list(set(month for room_data in usage.values() for month in room_data.keys())))

    # Header for the table-like output
    header_parts = ["  Month       "]
    for room in rooms_to_track:
        header_parts.append(f"{room} (Hours / % YoY)")
    print(" ".join(header_parts))
    print("-" * (14 + len(rooms_to_track) * 25)) # Adjust separator length

    for month_str in all_months:
        current_date = datetime.strptime(month_str, "%Y-%m")
        previous_year_date = current_date.replace(year=current_date.year - 1)
        previous_year_month_str = previous_year_date.strftime("%Y-%m")

        line_parts = [f"  {month_str}:"]

        for room in rooms_to_track:
            current_hours = usage.get(room, {}).get(month_str, 0.0)
            previous_year_hours = usage.get(room, {}).get(previous_year_month_str, 0.0)

            diff_str = ""
            # Only calculate percentage if there's data for the previous year
            if previous_year_hours > 0:
                diff = current_hours - previous_year_hours
                percent = (diff / previous_year_hours * 100)
                sign = "+" if percent >= 0 else ""
                diff_str = f" ({sign}{percent:.1f}%)"
            elif current_hours > 0 and previous_year_hours == 0:
                diff_str = " (New Data)" # Indicate no previous year data to compare
            else:
                diff_str = " " # No data for current or previous year

            # Format to align nicely
            line_parts.append(f"{current_hours:7.2f}h{diff_str:<15}") # Adjusted padding

        print(" ".join(line_parts))


if __name__ == "__main__":
    print(f"🔍 Tracking usage for rooms: {rooms_to_track} starting from {start_month}")
    room_map = load_room_to_external_ids(room_mapping_json)
    usage_data = build_monthly_usage(csv_file, room_map, start_month, rooms_to_track)
    print_monthly_usage_with_diff(usage_data, rooms_to_track)
