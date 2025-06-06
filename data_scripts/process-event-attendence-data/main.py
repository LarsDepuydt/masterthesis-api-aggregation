import json
from collections import defaultdict

# Load the JSON file
with open("getEventsOnCertainDay.json", "r", encoding="utf-8") as f:
    data = json.load(f)

event_attendance = defaultdict(int)

# Traverse buildings -> floors -> rooms -> events
for building in data["data"]["buildings"]:
    for floor in building.get("floors", []):
        for room in floor.get("rooms", []):
            for event in room.get("events", []):
                subject = event.get("subject", "Unknown Event")
                form_participants = event.get("formParticipants")
                if form_participants is not None:
                    event_attendance[subject] += form_participants
                else:
                    breakdown = event.get("departmentBreakdown", [])
                    attendee_sum = sum(d.get("attendeeCount", 0) for d in breakdown)
                    event_attendance[subject] += attendee_sum

# Filter out events with zero attendance and sort
sorted_attendance = sorted(
    ((subject, count) for subject, count in event_attendance.items() if count > 0),
    key=lambda x: x[1],
    reverse=True
)

# Print results
for subject, count in sorted_attendance:
    print(f"{subject}: {count} attendees")
