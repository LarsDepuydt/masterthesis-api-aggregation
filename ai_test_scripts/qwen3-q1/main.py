import requests
import json
from datetime import datetime

# Configuration
GRAPHQL_ENDPOINT = "http://localhost:4000/graphql"  # Replace with actual endpoint
HEADERS = {}

def fetch_telemetry_data(entrance_id: str, start_time: datetime, end_time: datetime = None):
    """
    Fetch telemetry data from a specific entrance within a time window.

    Args:
        entrance_id: The ID of the entrance to query
        start_time: Start of time window (datetime object)
        end_time: End of time window (datetime object, optional)

    Returns:
        dict: GraphQL response or error information
    """
    # Convert datetime objects to ISO format strings
    start_iso = start_time.isoformat() + "Z"  # Ensure UTC format
    end_iso = end_time.isoformat() + "Z" if end_time else None

    # GraphQL query with variables
    query = """
    query ($ids: [ID!]!, $startTime: Time!, $endTime: Time) {
        getEntrances(ids: $ids) {
        telemetryData(startTime: $startTime, endTime: $endTime) {
            timestamp
            value
        }
        }
    }
    """

    # Variables for the query
    variables = {
        "ids": [entrance_id],
        "startTime": start_iso,
        "endTime": end_iso
    }

    # Make the request
    try:
        response = requests.post(
            GRAPHQL_ENDPOINT,
            headers=HEADERS,
            json={"query": query, "variables": variables}
        )
        response.raise_for_status()  # Raise HTTP errors
        return response.json()

    except requests.exceptions.RequestException as e:
        return {"error": f"Request failed: {str(e)}"}
    except json.JSONDecodeError:
        return {"error": "Failed to parse response JSON"}

# Example usage
if __name__ == "__main__":
    # Parse timestamps from strings (or use direct datetime objects)
    start = datetime.fromisoformat("2023-10-01T08:00:00")
    end = datetime.fromisoformat("2023-10-01T10:00:00")

    result = fetch_telemetry_data("1234", start, end)

    # Print results or error
    print(json.dumps(result, indent=2))
