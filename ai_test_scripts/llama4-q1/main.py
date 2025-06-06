import datetime import requests

def get_current_time():
    response = time_server__get_current_time(timezone="Europe/Copenhagen")
    return response

def introspect_schema():
    response = mcp-graphql-aau-innovate-data-server__introspect-schema(argument="")
    return response

def query_graphql(query):
    response = mcp-graphql-aau-innovate-data-server__query-graphql(query=query)
    return response

def get_busiest_days(): current_time = get_current_time() current_year =
datetime.datetime.strptime(current_time, "%Y-%m-%dT%H:%M:%S.%fZ").year
start_time = f"{current_year}-01-01T00:00:00Z"

query = f"""
query {{
    buildings(name: "TMV25") {{
    entrances {{
        telemetryData(startTime: "{start_time}") {{
        timestamp
        peopleCount
        }}
    }}
    }}
}}
"""

response = query_graphql(query)
data = response['data']['buildings'][0]['entrances']

daily_visitors = {}
for entrance in data:
    for telemetry in entrance['telemetryData']:
        date = datetime.datetime.strptime(telemetry['timestamp'], "%Y-%m-
%dT%H:%M:%S.%fZ").date()
        daily_visitors[date] = daily_visitors.get(date, 0) +
telemetry['peopleCount']

sorted_days = sorted(daily_visitors.items(), key=lambda x: x[1],
reverse=True)
busiest_days = sorted_days[:10]

return busiest_days

busiest_days = get_busiest_days() for day, visitors in busiest_days:
print(f"{day}: {visitors}")
