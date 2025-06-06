import requests
from collections import defaultdict
from datetime import datetime, timedelta
import os
import json

GRAPHQL_ENDPOINT = "http://130.225.37.89:4000/graphql"
API_HEADERS = {}

def get_room_sensor_mapping():
    """Fetch mapping of rooms to their sensors' external IDs"""
    query = """
    query {
      buildings {
        floors {
          rooms {
            name
            sensors {
              externalID
            }
          }
        }
      }
    }
    """
    response = requests.post(
        GRAPHQL_ENDPOINT,
        headers=API_HEADERS,
        json={"query": query}
    )

    if response.status_code != 200:
        raise Exception(f"GraphQL query failed: {response.status_code}, {response.text}")

    return [
        room
        for building in response.json()["data"]["buildings"]
        for floor in building["floors"]
        for room in floor["rooms"]
    ]


TREND_DATA_DIR = "trenddata"

def process_trend_data(external_id, year):
    """Process all monthly trenddata files for a sensor in the given year"""
    events = []

    for month in range(1, 13):
        month_str = f"{month:02d}"
        filename = f"{external_id}_{year}-{month_str}.json"
        filepath = os.path.join(TREND_DATA_DIR, filename)

        if not os.path.exists(filepath):
            continue

        try:
            with open(filepath, 'r') as f:
                data = json.load(f)
                # Filter valid events and parse timestamps
                valid_events = []
                for entry in data:
                    try:
                        timestamp = datetime.fromisoformat(entry['timestamp'])
                        value = float(entry['value'])
                        valid_events.append({'timestamp': timestamp, 'value': value})
                    except (KeyError, ValueError) as e:
                        print(f"Skipping invalid entry in {filename}: {e}")

                # Sort by timestamp
                events.extend(sorted(valid_events, key=lambda x: x['timestamp']))

        except (IOError, json.JSONDecodeError) as e:
            print(f"Error reading {filename}: {e}")

    return calculate_active_duration(events)


def calculate_active_duration(events):
    """Calculate total duration where value was active (1.0)"""
    total_duration = timedelta()

    # Sort events by timestamp
    events.sort(key=lambda x: x['timestamp'])

    # Track previous event
    prev_event = None

    for event in events:
        # Skip invalid states
        if event['value'] not in (0.0, 1.0):
            print(f"Skipping event with invalid value: {event['value']}")
            continue

        if prev_event is not None:
            # Check for state transition to active (1.0)
            if event['value'] == 1.0 and prev_event['value'] == 0.0:
                # Calculate time between state transitions
                duration = event['timestamp'] - prev_event['timestamp']
                if duration.total_seconds() > 0:  # Only count forward in time
                    total_duration += duration

        prev_event = event

    # Convert to hours
    return total_duration.total_seconds() / 3600  # Convert seconds to hours


def analyze_room_activity(room_sensor_mapping, year):
    """Aggregate sensor values per room for the given year"""
    room_activity = defaultdict(float)

    for room in room_sensor_mapping:
        room_name = room.get('name', 'Unnamed Room')
        for sensor in room.get('sensors', []):
            external_id = sensor.get('externalID')
            if external_id:
                room_activity[room_name] += process_trend_data(external_id, year)

    # Sort by total activity (descending)
    return sorted(room_activity.items(), key=lambda x: x[1], reverse=True)


def main():
    target_year = 2023  # Adjust as needed

    try:
        room_sensor_mapping = get_room_sensor_mapping()
        results = analyze_room_activity(room_sensor_mapping, target_year)

        print(f"\nRoom Activity Summary for {target_year} (in hours):")
        for room, total in results:
            print(f"{room}: {total:.2f} hours")
    except Exception as e:
        print(f"An error occurred: {e}")

main()
