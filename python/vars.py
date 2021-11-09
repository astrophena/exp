import requests

r = requests.get("https://bot.astrophena.name/debug/vars")

json = r.json()
raw = r.raw

print(f"JSON: {json}")
print(f"Raw: {raw}")
