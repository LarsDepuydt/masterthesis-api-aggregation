import requests
from datetime import datetime, timedelta
import json

# Define the GraphQL endpoint
GRAPHQL_ENDPOINT = "http://130.225.37.89:4000/graphql"

# Step 1: Get the current year
current_year = datetime.now().year

# Query to fetch telemetry data for all entrances of each building starting from the beginning of the current year
telemetry_query = """
{
    buildings {
    entrances {
        id
        telemetryData(startDate: "%sT00:00:00Z", endDate: "%sT23:59:59Z") {
        timestamp
        value
        }
    }
    }
}
""" % (current_year, current_year)
# Function to query the GraphQL endpoint
def run_query(query):
    response = requests.post(GRAPHQL_ENDPOINT, json={'query': query})
    if response.status_code == 200:
        return response.json()
    else:
        raise Exception("Query failed to run with return code {}".format(response.status_code))

# Run the telemetry query
entrances_data = run_query(telemetry_query)

# Process the nested structure returned by the query, summing the number of visitors for each hour and aggregating them by day
visitor_counts = {}
for building in entrances_data['data']['buildings']:
    if building['name'] == 'TM51':  # Only process data for TM51 building
        for entrance in building['entrances']:
            for data_point in entrance['telemetryData']:
                timestamp = datetime.fromisoformat(data_point['timestamp'].replace('Z', '+00:00'))
                date_str = timestamp.date().isoformat()
                if date_str not in visitor_counts:
                    visitor_counts[date_str] = 0
                visitor_counts[date_str] += data_point['visitorCount']

# Sort the days by the number of visitors in descending order and get the top 10 busiest days
sorted_days = sorted(visitor_counts.items(), key=lambda item: item[1], reverse=True)[:10]

# Display the top 10 busiest days
print("Top 10 Busiest Days in TM51 Building for {}:".format(current_year))
for day, count in sorted_days:
    print(f"{day}: {count} visitors")
