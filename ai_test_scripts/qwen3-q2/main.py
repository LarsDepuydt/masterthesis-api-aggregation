import requests
from datetime import date, datetime, timedelta, time, timezone
from collections import defaultdict

# Configuration
GRAPHQL_URL = 'http://130.225.37.89:4000/graphql'
START_DATE = datetime(2022, 1, 1, tzinfo=timezone.utc)
END_DATE = datetime(2025, 5, 22, tzinfo=timezone.utc)
DAILY_WINDOW_START = time(8, 0)
DAILY_WINDOW_END = time(16, 0)
BUILDING_ID = "TMV25"

def run_graphql_query(query):
    """Send a GraphQL query to the API and return the JSON response."""
    response = requests.post(GRAPHQL_URL, json={'query': query})
    response.raise_for_status()
    return response.json()

def fetch_rooms_and_events(start, end):
    """Fetch rooms (via floors) and their events from the GraphQL API."""
    start_iso = start.astimezone(timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    end_iso = end.astimezone(timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')

    query = f"""
    query {{
        buildings(ids: ["{BUILDING_ID}"]) {{
            id
            floors {{
                id
                rooms {{
                    id
                    name
                    type
                    events(startTime: "{start_iso}", endTime: "{end_iso}") {{
                        eventID
                        start
                        end
                    }}
                }}
            }}
        }}
    }}
    """
    print("Executing query:", query)
    data = run_graphql_query(query)

    # Extract all rooms from all floors of the specified building
    rooms = []
    for building in data['data']['buildings']:
        for floor in building['floors']:
            rooms.extend(floor['rooms'])

    return rooms

def calculate_daily_utilization(events):
    """Process events and compute daily utilization for a room."""
    usages = defaultdict(timedelta)

    for event in events:
        try:
            start_dt = datetime.fromisoformat(event['start'])
            end_dt = datetime.fromisoformat(event['end'])
        except (ValueError, KeyError):
            continue  # Skip invalid events

        current_date = start_dt.date()
        end_event_date = end_dt.date()

        while current_date <= end_event_date:
            day_start = datetime.combine(current_date, DAILY_WINDOW_START).replace(tzinfo=timezone.utc)
            day_end = datetime.combine(current_date, DAILY_WINDOW_END).replace(tzinfo=timezone.utc)

            window_start = max(start_dt, day_start)
            window_end = min(end_dt, day_end)

            if window_start < window_end:
                usages[current_date] += (window_end - window_start)

            current_date += timedelta(days=1)

    return usages

def analyze_utilization(rooms_data):
    """Analyze utilization across all rooms in the building."""
    all_daily_usages = defaultdict(timedelta)
    peak_usage = timedelta(0)
    peak_room = None
    full_utilization = False

    for room in rooms_data:
        room_id = room['id']
        room_events = room['events']
        usages = calculate_daily_utilization(room_events)

        for date, usage in usages.items():
            all_daily_usages[date] += usage
            if usage > peak_usage:
                peak_usage = usage
                peak_room = room
            # Check for full utilization (8 hours)
            if usage.total_seconds() >= 28800:  # 8 hours in seconds
                full_utilization = True

    total_days = (END_DATE - START_DATE).days + 1

    total_usage = sum(all_daily_usages.values(), timedelta())
    average_daily_usage = total_usage.total_seconds() / (total_days * len(rooms_data))

    average_daily_usage_hours = average_daily_usage / 3600

    return {
        'average_daily_usage_hours': round(average_daily_usage_hours, 2),
        'peak_usage_hours': round(peak_usage.total_seconds() / 3600, 2),
        'peak_room_name': peak_room['name'] if peak_room else 'N/A',
        'full_utilization': full_utilization,
        'total_meeting_rooms': len(rooms_data)
    }

# Example usage
if __name__ == "__main__":
    rooms = fetch_rooms_and_events(START_DATE, END_DATE)
    results = analyze_utilization(rooms)

    print("\nUtilization Analysis Results:")
    print(f"Total meeting rooms analyzed: {results['total_meeting_rooms']}")
    print(f"Average daily utilization: {results['average_daily_usage_hours']} hours")
    print(f"Peak room utilization: {results['peak_usage_hours']} hours")
    print(f"Peak room name: {results['peak_room_name']}")
    print(f"Any room at full capacity (8h+): {'Yes' if results['full_utilization'] else 'No'}")
