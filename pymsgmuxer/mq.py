"""Using asyncio to simulate message queue"""

import os
import asyncio
from datetime import datetime
from datetime import timedelta
from dataclasses import dataclass

class MsgNode:
    '''Message Node'''
    def __init__(self, msgid, msg):
        self.id = msgid
        self.ts = int(datetime.utcnow().timestamp() * 1000000)
        self.data = msg
        self.next = None

    def __repr__(self):
        return f'<{type(self).__name__}: {self.id}, {self.ts} => {self.data}>'

    def __eq__(self, node):
        if not node:
            return False
        return self.id == node.id and self.ts == node.ts and self.data == node.data

@dataclass
class ConsumerInfo:
    '''Consumer Information Tracking'''
    pcount: int
    lastnode: MsgNode

class MsgQueue:
    '''Message Queue'''

    begin_mark = "BeginMarker"
    end_mark = "EndMarker"

    def __init__(self, retention, cleanfreq):
        self._cinfo = {} # consumer info tracker

        self._retention = retention
        self._cleanfreq = cleanfreq

        # Using Begin and End Marker
        self._lhead = MsgNode(msgid=0, msg=MsgQueue.begin_mark)
        self._ltail = MsgNode(msgid=0, msg=MsgQueue.end_mark)
        self._lhead.next = self._ltail
        self._linsp = self._lhead

        self._seqid = 0
        self._currcnt = 0

    def put(self, ident, msg):
        '''method to put item in q'''
        msg = "[{}] {}".format(self._seqid, msg)
        node = MsgNode(msgid=ident, msg=msg)

        # The insp (insert pointer) tracks insertion point
        node.next = self._linsp.next
        self._linsp.next = node
        self._linsp = node

        self._seqid += 1
        self._currcnt += 1

        return True

    def get(self, cid):
        '''method to get item from q'''
        retmsg = None

        # Get offset dict for cid
        node = self._cinfo.get(cid, None)
        if not node:
            return retmsg

        # Get to be visiting node for cid
        node = self._cinfo[cid].lastnode
        node = node.next
        if not node or node.data == MsgQueue.end_mark:
            return retmsg

        retmsg = node.data

        # Track the consumed node
        self._cinfo[cid].lastnode = node
        self._cinfo[cid].pcount += 1
        return retmsg

    def register_consumer(self, cid):
        '''add consumer using the queue'''
        if cid in self._cinfo:
            print(f'consumer {cid} already registered, exception/error')
            return False

        # Initialize dict to store list head and count processed
        self._cinfo[cid] = ConsumerInfo(pcount=0, lastnode=self._lhead)
        return True

    def unregister_consumer(self, cid):
        '''remove consumer using the queue'''
        if cid in self._cinfo:
            del self._cinfo[cid]

    def _is_consumed(self, dnode):
        '''check if all consumers have consumed dnode'''
        cstatus = True
        cids = []
        for ckey, cdict in self._cinfo.items():
            cnode = cdict.lastnode
            if dnode == cnode:
                cstatus = False
                cids.append(ckey)
        return cstatus, cids

    def _print_status(self):
        result = "Msg Queue Status:\n"
        result += f'total = {self._seqid}, current = {self._currcnt}\n'
        #result += f'lhead: {self._lhead}, ltail: {self._ltail}\n'
        # walker = self._lhead.next
        # while walker.data != MsgQueue.end_mark:
            # result += f'  {walker!r}\n'
            # walker = walker.next
        result += f'consumer count = {len(self._cinfo.keys())}\n'
        # for ckey, cdict in self._cinfo.items():
        #     result += f'  {ckey}: count={cdict.pcount}, last-msg={cdict.lastnode!r}\n'
        print(result)

    async def cleanqueue(self):
        '''task to clean old messages'''
        procid = os.getpid()
        tstamp = datetime.utcnow().strftime("%a, %d %b %Y %H:%M:%S.%f")
        print(f'{tstamp}: Pid[{procid}] MsgQueue clean task begins...')
        while True:
            await asyncio.sleep(self._cleanfreq)
            # dbgmsg = ""
            rtime = datetime.utcnow() - timedelta(seconds=self._retention)
            rtimeint = int(rtime.timestamp() * 1000000)

            walker = self._lhead.next
            rcount = 0
            while walker.data != MsgQueue.end_mark and walker.ts < rtimeint:
                allconsume, cids = self._is_consumed(walker)
                if not allconsume:
                    # dbgmsg = f'{cids} is yet to consume {walker!r}'
                    break
                self._lhead.next = walker.next
                walker.next = None
                # print(f'Removing {walker!r}')
                walker = self._lhead.next
                self._currcnt -= 1
                rcount += 1

            # rtimestr = rtime.strftime("%a, %d %b %Y %H:%M:%S.%f")
            # if not dbgmsg:
            #     dbgmsg = f'Removing {rcount} nodes, retention time: {rtimestr}'
            # else:
            #     dbgmsg = f'Removing {rcount} nodes, retention time: {rtimestr}\n' + dbgmsg
            # print(dbgmsg)
            self._print_status()
