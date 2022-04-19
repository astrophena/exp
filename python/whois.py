import requests

r = requests.get('https://feed.coin-tone.ts.net/_proxy/whois')

json = r.json()

node = json['Node']['ComputedName']
user = json['UserProfile']['DisplayName']

print(f"Hello {user}! You're running this from {node}.")
