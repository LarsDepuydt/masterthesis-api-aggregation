
import json
from collections import defaultdict
from datetime import datetime

# Load the JSON data
with open("getEventsExternal.json", encoding="utf-8") as f:
    data = json.load(f)

# Dictionary to store external meetings per year
external_meetings_per_year = defaultdict(int)

# Traverse the nested structure
for building in data.get("data", {}).get("buildings", []):
    for floor in building.get("floors", []):
        for room in floor.get("rooms", []):
            for event in room.get("events", []):
                if any(participant.get("isExternal", False) for participant in event.get("departmentBreakdown", [])):
                    start_time = event.get("start")
                    if start_time:
                        try:
                            year = datetime.fromisoformat(start_time.replace("Z", "+00:00")).year
                            external_meetings_per_year[year] += 1
                        except ValueError:
                            pass  # skip if the date format is invalid

# Print the results
for year in sorted(external_meetings_per_year):
    print(f"{year}: {external_meetings_per_year[year]} meeting(s) with external participants")
