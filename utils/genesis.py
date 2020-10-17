import json

genesis=None
import sys
with open(sys.argv[1], "r") as f:
    genesis=json.load(f)
def insertGensisNodeRewards():
    prefix_query="insert into genesis(NodeID, Timelock, Wits) values"
    timeSorted={}
    l=[]
    ml =0
    for addrs in genesis['alloc']:
        for addr in addrs:
            timelock=addr['timelock']
            val=addr['value']
            addr=addr['address']
            if addr not in timeSorted.keys():
                timeSorted[addr] = {"t":[timelock], "v": [val]}
            else:
                timeSorted[addr]["t"].append(timelock)
                timeSorted[addr]["v"].append(val)
    for k,v in timeSorted.items():
        a="('%s','%s', '%s')" %(k, ",".join(v["t"]), ",".join(v['v']))
        ml = max(ml , len(",".join(v['t']) ))
        ml = max(ml , len(",".join(v['v']) ))
        l.append(a)
    print(prefix_query + ",".join(l)+";")

def unlockTimeGroupQuery():
    timelocks = set()
    for addrs in genesis['alloc']:
        for addr in addrs:
            timelock=addr['timelock']
            val=addr['value']
            addr=addr['address']
            timelocks.add(timelock)
    for t in timelocks:
        q = "insert into genesisUnlockNotify(notifyTimeStamp, notified) values (%s, false);" % t
        print(q)

def changeAddrForMainnet():
    mainnetAddrs = set()
    for addrs in genesis['alloc']:
        for addr in addrs:
            mainnetAddrs.add(addr['address'])
    for addr in mainnetAddrs:
        q = "update userNodeMap set NodeID='%s' where NodeID like 't%s" % (addr,addr[:-10])
        print(q+"%';")

changeAddrForMainnet()
# insertGensisNodeRewards()
# unlockTimeGroupQuery()