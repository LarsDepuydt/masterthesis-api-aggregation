
import requests
from datetime import datetime, timedelta
import os
import json
from time import sleep

# Authentication
USERNAME = ""
PASSWORD = ""

# Endpoint
BASE_URL = "https://bms-api.build.aau.dk/api/v1/trenddata"

# External IDs
external_ids = [
    "176180", "176144", "176135", "176126", "176117", "176108", "176099", "847161", "175882", "176342",
    "176333", "176324", "176315", "235564", "176306", "176297", "285405", "176288", "176279", "176270",
    "176261", "176360", "176351", "295938", "295947", "176252", "176243", "176234", "176225", "176216",
    "176207", "176198", "176189", "176857", "176848", "176839", "176830", "295956", "176821", "176812",
    "176803", "176794", "176776", "176767", "176758", "285839", "176749", "285840", "176740", "176731",
    "688816", "176721", "176712", "176703", "176694", "176685", "176676", "176667", "176658", "176649",
    "176640", "176631", "176622", "176613", "176603", "176594", "176585", "176576", "176567", "176558",
    "179929", "179938", "179947", "363259", "295974", "179956", "179965", "179983", "179974", "688099",
    "363249", "295965", "363239", "363273", "179992", "180001", "180010", "180046", "363229", "180019",
    "180028", "180037", "235558"
]

# Time range
start_date = datetime(2022, 9, 1)
end_date = datetime(2025, 5, 20)

# Output folder
output_folder = "trenddata"
os.makedirs(output_folder, exist_ok=True)

# Format ISO time
def iso(dt):
    return dt.strftime("%Y-%m-%dT%H:%M:%SZ")

# Loop through all months and IDs
for external_id in external_ids:
    current = start_date
    while current < end_date:
        # Calculate next month's first day
        next_month = (current.replace(day=28) + timedelta(days=4)).replace(day=1)
        start_str = iso(current)
        end_str = iso(min(next_month, end_date))
        month_str = current.strftime('%Y-%m')

        # Construct filename
        filename = os.path.join(output_folder, f"{external_id}_{month_str}.json")

        if os.path.exists(filename):
            print(f"Skipping {filename}, already exists.")
        else:
            print(f"Fetching ID {external_id} from {start_str} to {end_str}...")
            try:
                response = requests.get(BASE_URL, params={
                    "externallogid": external_id,
                    "starttime": start_str,
                    "endtime": end_str
                }, auth=(USERNAME, PASSWORD))
                response.raise_for_status()
                data = response.json()

                # Save data
                with open(filename, "w", encoding="utf-8") as f:
                    json.dump(data, f, ensure_ascii=False, indent=2)

            except requests.HTTPError as e:
                print(f"HTTP error for {external_id} from {start_str} to {end_str}: {e}")
            except Exception as e:
                print(f"Unexpected error: {e}")

            sleep(0.2)

        current = next_month
