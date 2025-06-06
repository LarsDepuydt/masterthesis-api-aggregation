
from statistics import mean, median
from datetime import timedelta

# Helper function to convert time strings to seconds
def time_to_seconds(t):
    if 'm' in t:
        parts = t.split('m')
        minutes = int(parts[0])
        seconds = int(parts[1].replace('s', '')) if len(parts) > 1 else 0
    else:
        minutes = 0
        seconds = int(t.replace('s', ''))
    return minutes * 60 + seconds

# Input: raw timing data
qwen3_times = [
    "2m44s", "1m51s", "1m19s", "18m13s", "5m1s", "2m58s", "3m56s",
    "2m34s", "5m18s", "4m41s", "5m35s", "8m7s"
]
mistral_times = ["13m16s", "4m31s", "54s", "1m4s"]
llama4_times = ["4m25s", "57s", "1m16s", "28s"]
watt_tool_times = ["4m21s", "1m11s", "47s", "28s"]

# Convert all to seconds
data = {
    "Qwen3": [time_to_seconds(t) for t in qwen3_times],
    "Mistral": [time_to_seconds(t) for t in mistral_times],
    "Llama 4": [time_to_seconds(t) for t in llama4_times],
    "Watt Tool": [time_to_seconds(t) for t in watt_tool_times],
}

# Calculate average and median, and print results
for model, times in data.items():
    avg = mean(times)
    med = median(times)
    print(f"{model}:")
    print(f"  Average: {str(timedelta(seconds=int(avg)))}")
    print(f"  Median:  {str(timedelta(seconds=int(med)))}\n")
