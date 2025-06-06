import json

def extract_sensor_ids(json_file_path, rooms_to_track):
    with open(json_file_path, 'r') as file:
        data = json.load(file)

    sensor_map = {room: [] for room in rooms_to_track}

    for building in data.get("data", {}).get("buildings", []):
        for floor in building.get("floors", []):
            for room in floor.get("rooms", []):
                room_number = room.get("roomNumber")
                if room_number in rooms_to_track:
                    for sensor in room.get("sensors", []):
                        sensor_map[room_number].append(sensor.get("externalID"))

    return sensor_map

# Example usage
rooms_to_track = ["A.101a", "A.101b"]
result = extract_sensor_ids("getRoomActivity.json", rooms_to_track)

print("Sensor IDs in tracked rooms:")
print(json.dumps(result, indent=2))
