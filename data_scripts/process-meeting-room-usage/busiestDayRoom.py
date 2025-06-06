
import json
from datetime import datetime, timezone
from collections import defaultdict

# Constants
WORKDAY_MINUTES = 8 * 60  # 480 minutes per workday

# Load JSON file
with open("getAllEvents.json", "r", encoding="utf-8") as f:
    data = json.load(f)

# Storage: room_id -> date -> total minutes
room_daily_minutes = defaultdict(lambda: defaultdict(int))

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
                room_daily_minutes[room_id][date_key] += duration_minutes

# Output the results
for room_id, usage_by_day in room_daily_minutes.items():
    print(f"\nRoom: {room_id}")
    # Sort days by most occupied first
    sorted_days = sorted(usage_by_day.items(), key=lambda x: x[1], reverse=True)
    for date, minutes in sorted_days:
        pct = (minutes / WORKDAY_MINUTES) * 100
        print(f"{date}: {pct:.1f}% occupied ({minutes} min)")
