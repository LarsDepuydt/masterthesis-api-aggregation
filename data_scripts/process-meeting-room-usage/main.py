import json
from datetime import datetime, timedelta, timezone
from collections import defaultdict
from calendar import monthrange

# Load the JSON file
with open("getAllEvents.json", "r", encoding="utf-8") as f:
    data = json.load(f)

# Constants
WORKDAY_MINUTES = 8 * 60
START_DATE = datetime(2022, 1, 1, tzinfo=timezone.utc)

# Helper to get half-year key from a date
def get_half_year_key(date: datetime):
    return f"{date.year}-H1" if date.month <= 6 else f"{date.year}-H2"

# occupancy_by_interval_and_room[interval][room_id][date] = total_minutes
occupancy_by_interval_and_room = defaultdict(lambda: defaultdict(lambda: defaultdict(int)))
interval_latest_date = defaultdict(lambda: START_DATE)

# Traverse all rooms and events
for building in data["data"]["buildings"]:
    for floor in building.get("floors", []):
        for room in floor.get("rooms", []):
            room_id = room.get("id", "Unknown Room")
            for event in room.get("events", []):
                start_str = event.get("start")
                duration_minutes = event.get("durationMinutes")
                if not duration_minutes or not start_str:
                    continue

                start_time = datetime.fromisoformat(start_str.replace("Z", "+00:00"))
                date_key = start_time.date().isoformat()
                interval = get_half_year_key(start_time)

                occupancy_by_interval_and_room[interval][room_id][date_key] += duration_minutes

                if start_time > interval_latest_date[interval]:
                    interval_latest_date[interval] = start_time

# Compute all weekday dates per interval
interval_weekdays = {}
for interval, latest_dt in interval_latest_date.items():
    year, half = map(str, interval.split("-"))
    year = int(year)
    start = datetime(year, 1 if half == "H1" else 7, 1, tzinfo=timezone.utc)
    end_month = 6 if half == "H1" else 12
    end_day = monthrange(year, end_month)[1]
    end = datetime(year, end_month, end_day, tzinfo=timezone.utc)

    weekdays = set()
    current = start
    while current <= end and current <= latest_dt:
        if current.weekday() < 5:  # Mon–Fri
            weekdays.add(current.date().isoformat())
        current += timedelta(days=1)

    interval_weekdays[interval] = weekdays

# Compute and print averages
for interval in sorted(occupancy_by_interval_and_room.keys()):
    room_usages = {}
    total_pct_sum = 0
    num_rooms = 0
    total_days = len(interval_weekdays[interval])

    print(f"\n--- {interval} ---")

    for room_id, usage_by_day in occupancy_by_interval_and_room[interval].items():
        total_minutes = 0
        for day in interval_weekdays[interval]:
            total_minutes += usage_by_day.get(day, 0)

        avg_pct = (total_minutes / (total_days * WORKDAY_MINUTES)) * 100
        room_usages[room_id] = avg_pct

        total_pct_sum += avg_pct
        num_rooms += 1

    # Sort and display per-room usage
    sorted_usage = sorted(room_usages.items(), key=lambda x: x[1], reverse=True)
    for room_id, pct in sorted_usage:
        print(f"{room_id}: {pct:.2f}% average usage")

    # Print form-wide average
    if num_rooms > 0:
        print(f"\nForm-wide average for {interval}: {(total_pct_sum / num_rooms):.2f}%")
    else:
        print(f"\nForm-wide average for {interval}: No data")
