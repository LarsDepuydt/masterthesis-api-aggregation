import requests
import json

n = "1544"
starttime = "2025-03-17T00:00:00.000"
endtime = "2025-04-07T00:00:00.000"

# Load token from the text file
token = ""
try:
    with open("token", "r", encoding="utf-8") as f:
        token = f.read().strip()  # Strip to remove any leading/trailing whitespace
        print("Token loaded")
except FileNotFoundError:
    print("Token file not found")
except Exception as e:
    print(f"Error reading token file: {e}")

url = f"https://outlook.office.com/owa/service.svc?action=GetCalendarView&app=Calendar&n={n}"

headers = {
    "User-Agent": "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0",
    "Accept": "*/*",
    "Accept-Language": "en-US,en;q=0.5",
    "action": "GetCalendarView",
    "content-type": "application/json; charset=utf-8",
    "ms-cv": "zyu7vzCUKv+PnvKg7aO2bn.1544",
    "prefer": "exchange.behavior=\"IncludeThirdPartyOnlineMeetingProviders\"",
    "x-anchormailbox": "PUID:10032003F2A4A94B@f5dbba49-ce06-496f-ac3e-0cf14361d934",
    "x-owa-correlationid": "b24fdeeb-b5f7-e543-baa1-075d4596197c",
    "x-owa-hosted-ux": "false",
    "x-owa-sessionid": "f687c167-de15-4d8c-85b6-acbda91bdf68",
    "x-req-source": "Calendar",
    "Sec-Fetch-Dest": "empty",
    "Sec-Fetch-Mode": "cors",
    "Sec-Fetch-Site": "same-origin",
    "authorization": f"Bearer {token}",
}

data = {
    "__type": "GetCalendarViewJsonRequest:#Exchange",
    "Header": {
        "__type": "JsonRequestHeaders:#Exchange",
        "RequestServerVersion": "V2018_01_08",
        "TimeZoneContext": {
            "__type": "TimeZoneContext:#Exchange",
            "TimeZoneDefinition": {
                "__type": "TimeZoneDefinitionType:#Exchange",
                "Id": "Romance Standard Time"
            }
        }
    },
    "Body": {
        "__type": "GetCalendarViewRequest:#Exchange",
        "CalendarId": {
            "__type": "TargetFolderId:#Exchange",
            "BaseFolderId": {
                "__type": "FolderId:#Exchange",
                "Id": "AAMkADVhMzMwMTQzLTlkMjMtNDRkNS04YWYyLTIyYTFmMDg4ODZhZgAuAAAAAADeY/v4vOW9QITYM7urugEqAQAF4KhP8w1MQ7XGRvsFkWRZAABRq5cCAAA="
            }
        },
        "RangeStart": starttime,
        "RangeEnd": endtime,
        "ClientSupportsIrm": True,
        "OptimizeExtendedPropertyLoading": True
    }
}

response = requests.post(url, headers=headers, json=data)

# Print status code and response
print(response.status_code)

# Save response JSON to a file
with open("response.json", "w", encoding="utf-8") as f:
    json.dump(response.json(), f, indent=4)

print("Response saved to response.json")
