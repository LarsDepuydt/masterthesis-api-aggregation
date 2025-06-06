
import json
from datetime import datetime, timedelta, timezone
from collections import defaultdict

# Load JSON file
with open("getAllEvents.json", "r", encoding="utf-8") as f:
    data = json.load(f)

# Parameter: How many top results to show
top_x = 20  # <- You can change this to any number you want

# Store: (date, time window) -> set of room_ids
window_room_usage = defaultdict(set)

# Valid rooms that are considered meeting rooms (have email set)
meeting_rooms = set()

# Sliding window parameters
WINDOW_SIZE = timedelta(hours=1)
STEP = timedelta(minutes=15)  # For overlap detection

# Parse events and identify valid meeting rooms
for building in data["data"]["buildings"]:
    for floor in building.get("floors", []):
        for room in floor.get("rooms", []):
            room_id = room.get("id", "Unknown Room")
            email = room.get("email", "").strip()
            if not email:
                continue  # Skip non-meeting rooms

            meeting_rooms.add(room_id)

            for event in room.get("events", []):
                start_str = event.get("start")
                end_str = event.get("end")
                if not start_str or not end_str:
                    continue

                start = datetime.fromisoformat(start_str.replace("Z", "+00:00"))
                end = datetime.fromisoformat(end_str.replace("Z", "+00:00"))

                # Sliding windows from event.start - 1h to event.end
                window_start = start - WINDOW_SIZE + STEP
                while window_start + WINDOW_SIZE <= end:
                    window_end = window_start + WINDOW_SIZE
                    date_key = window_start.date().isoformat()
                    time_key = f"{window_start.time().strftime('%H:%M')}–{window_end.time().strftime('%H:%M')}"
                    key = (date_key, time_key)
                    window_room_usage[key].add(room_id)
                    window_start += STEP

# Count room usage in each window
usage_counts = [(key, len(rooms)) for key, rooms in window_room_usage.items()]
top_windows = sorted(usage_counts, key=lambda x: x[1], reverse=True)[:top_x]

# Print results
print(f"\nTop {top_x} busiest 1-hour periods:")
for (date, time_range), count in top_windows:
    print(f"{date} {time_range}: {count} rooms occupied out of {len(meeting_rooms)}")
