
import json
from datetime import datetime, date
from collections import defaultdict

def main():
    # Load JSON data
    with open("GetBuildingIDQuery.json", "r") as f:
        data = json.load(f)

    # Filter threshold date
    filter_start_date = date(2025, 1, 1)

    # Map entrance index to labels
    entrance_labels = {
        0: "A",  # C1: Indgang A
        1: "B",  # C3: Indgang B
        2: "C"   # C2: Indgang C
    }

    # Prepare storage
    daily_totals = defaultdict(lambda: defaultdict(int))  # {date: {entrance_label: sum}}

    # Process each entrance
    for idx, entrance in enumerate(data["data"]["buildings"][0]["entrances"]):
        label = entrance_labels.get(idx, f"Unknown_{idx}")
        for entry in entrance["telemetryData"]:
            entry_date = datetime.fromisoformat(entry["timestamp"].replace("Z", "+00:00")).date()
            if entry_date >= filter_start_date:
                date_str = entry_date.isoformat()
                daily_totals[date_str][label] += entry["value"]

    # Build final summary
    final_summary = {}
    for day, entrances in daily_totals.items():
        day_total = sum(entrances.values())
        final_summary[day] = {
            "per_entrance": entrances,
            "total": day_total
        }

    # Sort by total visitors descending
    sorted_days = sorted(final_summary.items(), key=lambda item: item[1]['total'], reverse=True)

    # Print the results
    for day, values in sorted_days:
        print(f"{day}: Total = {values['total']}, Per Entrance = {dict(values['per_entrance'])}")

if __name__ == "__main__":
    main()
