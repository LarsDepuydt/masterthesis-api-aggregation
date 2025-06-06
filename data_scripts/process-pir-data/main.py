
import json

# Load the JSON file
with open('Test.json', 'r', encoding='utf-8') as f:
    data = json.load(f)

# Extract the list of sensors
sensors = data.get('data', {}).get('sensors', [])

# Filter for PIR sensors with 'TM025' in path and without '30-dagstest'
pir_sensor_ids = [
    sensor['externalID']
    for sensor in sensors
    if 'pir' in sensor.get('sourcePath', '').lower()
    and 'tm025' in sensor.get('sourcePath', '').lower()
    and '30-dagstest' not in sensor.get('sourcePath', '').lower()
    and '/a/' not in sensor.get('sourcePath', '').lower()
]

# Print result as JSON array with double quotes
print(json.dumps(pir_sensor_ids))
