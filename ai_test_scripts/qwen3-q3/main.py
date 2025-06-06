import requests
from datetime import datetime
from collections import defaultdict

# Replace with your GraphQL API endpoint and authentication token
GRAPHQL_URL = "http://130.225.37.89:4000/graphql"

# Construct the updated query to include start time
query = """
{
    rooms {
    events(startTime: "2022-01-01T00:00:00Z") {
        start
        departmentBreakdown {
        isExternal
        }
    }
    }
}
"""

# Set up headers
headers = {
    "Content-Type": "application/json"
}

# Make the GraphQL request
response = requests.post(GRAPHQL_URL, headers=headers, json={"query":
query})

# Check for request errors
if response.status_code != 200:
    raise Exception(f"Query failed with status code {response.status_code}: {response.text}")

# Parse the response
data = response.json()

# Initialize a dictionary to store yearly counts
year_counts = defaultdict(int)

# Process the data
for room in data.get("data", {}).get("rooms", []):
    for event in room.get("events", []):
        start_time_str = event.get("start")
        if not start_time_str:
            continue  # Skip if no start time

        try:
            # Parse the ISO 8601 datetime string
            start_dt = datetime.fromisoformat(start_time_str.replace("Z",
"+00:00"))
            year = start_dt.year
        except ValueError:
            continue  # Skip invalid date formats

        # Check if any department in the event is external
        departments = event.get("departmentBreakdown", [])
        has_external = any(dept.get("isExternal", False) for dept in
departments)

        if has_external:
            year_counts[year] += 1

# Calculate the total
total = sum(year_counts.values())

# Output the results
print("Meetings with external partners by year:")
for year in sorted(year_counts):
    print(f"{year}: {year_counts[year]} meeting(s)")

print(f"\nTotal meetings with external partners: {total}")
