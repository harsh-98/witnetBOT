import json

genesis=None
import sys
with open(sys.argv[1], "r") as f:
    genesis=json.load(f)

addrs=genesis['alloc'][0]
for addr in addrs:
    q="insert into genesis(NodeID, Timelock, Wits) values ('%s', %s, %s);" %(addr['address'], addr['timelock'], addr['value'])
    print(q)
