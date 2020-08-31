"""Using asyncio to simulate message queue"""

import os
import random
import asyncio
from datetime import datetime
from datetime import timedelta
from dataclasses import dataclass
import aiofiles

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

class MsgChannel:
    '''Message Channel'''

    begin_mark = "BeginMarker"
    end_mark = "EndMarker"

    def __init__(self, retention, cleanfreq):
        self._cinfo = {} # consumer info tracker

        self._retention = retention
        self._cleanfreq = cleanfreq

        # Using Begin and End Marker
        self._lhead = MsgNode(msgid=0, msg=MsgChannel.begin_mark)
        self._ltail = MsgNode(msgid=0, msg=MsgChannel.end_mark)
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
        if not node or node.data == MsgChannel.end_mark:
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
        result = "Msg Channel Status:\n"
        result += f'total = {self._seqid}, current = {self._currcnt}\n'
        #result += f'lhead: {self._lhead}, ltail: {self._ltail}\n'
        walker = self._lhead.next
        while walker.data != MsgChannel.end_mark:
            # result += f'  {walker!r}\n'
            walker = walker.next
        result += f'consumer count = {len(self._cinfo.keys())}\n'
        for ckey, cdict in self._cinfo.items():
            result += f'  {ckey}: count={cdict.pcount}, last-msg={cdict.lastnode!r}\n'
        print(result)

    async def cleanqueue(self):
        '''task to clean old messages'''
        procid = os.getpid()
        tstamp = datetime.utcnow().strftime("%a, %d %b %Y %H:%M:%S.%f")
        print(f'{tstamp}: Pid[{procid}] MsgChannel clean task begins...')
        while True:
            await asyncio.sleep(self._cleanfreq)
            dbgmsg = ""
            rtime = datetime.utcnow() - timedelta(seconds=self._retention)
            rtimeint = int(rtime.timestamp() * 1000000)

            walker = self._lhead.next
            rcount = 0
            while walker.data != MsgChannel.end_mark and walker.ts < rtimeint:
                allconsume, cids = self._is_consumed(walker)
                if not allconsume:
                    dbgmsg = f'{cids} is yet to consume {walker!r}'
                    break
                self._lhead.next = walker.next
                walker.next = None
                # print(f'Removing {walker!r}')
                walker = self._lhead.next
                self._currcnt -= 1
                rcount += 1

            rtimestr = rtime.strftime("%a, %d %b %Y %H:%M:%S.%f")
            if not dbgmsg:
                dbgmsg = f'Removing {rcount} nodes, retention time: {rtimestr}'
            else:
                dbgmsg = f'Removing {rcount} nodes, retention time: {rtimestr}\n' + dbgmsg
            print(dbgmsg)
            self._print_status()

class Producer:
    '''Producer/Publisher module'''
    def __init__(self, instid, nmsgs, msgq, rwindow):
        self._instid = instid
        self._num_msgs = nmsgs
        self._mq = msgq
        self._rw = rwindow
        self._pname = "producer_{}".format(instid)

    def create_msg(self):
        '''Create message'''
        utctime = datetime.utcnow()
        utcns = int(utctime.timestamp() * 1000000)
        utcnsstr = utctime.strftime("%a, %d %b %Y %H:%M:%S.%f")
        msg = "hello from {} @ {}".format(self._pname, utcnsstr)
        return utcns, msg

    def put_msg(self, msgid, msg):
        '''Put message into queue'''
        rval = self._mq.put(msgid, msg)
        if not rval:
            print(f'{self._pname} put message failed id: {msgid}')

    async def process(self):
        '''Process function doing main work of producing messages'''
        procid = os.getpid()
        tstamp = datetime.utcnow().strftime("%a, %d %b %Y %H:%M:%S.%f")
        print(f'{tstamp}: Pid[{procid}] {self._pname} begins...')
        low, high = self._rw
        fname = "{}.txt".format(self._pname)
        fdesc = await aiofiles.open(fname, 'w')

        for _ in range(self._num_msgs):
            # Do processing and create message
            await asyncio.sleep(random.randint(low, high))

            msgid, msg = self.create_msg()
            await fdesc.write("{}: {}\n".format(msgid, msg))

            # Put message in the queue
            self.put_msg(msgid=msgid, msg=msg)

        fdesc.close()
        tstamp = datetime.utcnow().strftime("%a, %d %b %Y %H:%M:%S.%f")
        print(f'{tstamp}: Pid[{procid}] {self._pname} done, generated {self._num_msgs} msgs')

class Consumer:
    '''Consumer/Subscriber module'''
    def __init__(self, instid, msgq, rwindow):
        self._instid = instid
        self._mq = msgq
        self._rw = rwindow
        self._mcount = 0
        self._pname = "consumer_{}".format(instid)

    def get_msg(self):
        '''Consume message from queue'''
        msg = self._mq.get(self._pname)
        if msg:
            self._mcount += 1
        return msg

    async def process(self):
        '''The process() function doing the main work of consumer'''
        procid = os.getpid()
        low, high = self._rw

        # Register with the queue
        tstamp = datetime.utcnow().strftime("%a, %d %b %Y %H:%M:%S.%f")
        print(f'{tstamp}: Pid[{procid}] {self._pname} begins...')
        self._mq.register_consumer(self._pname)

        fname = "{}.txt".format(self._pname)
        fdesc = await aiofiles.open(fname, 'w')

        # work loop
        while True:
            msg = self.get_msg()
            if msg:
                await fdesc.write("{} {}\n".format(self._mcount, msg))

            # Do processing of the message
            await asyncio.sleep(random.randint(low, high))

async def aio_sessions(loop):
    '''Create session to concurrently fetch tickers.'''
    procid = os.getpid()
    print(f'Pid[{procid}]: asyncio session...')

    # Create message queue, producers and consumers
    mqueue = MsgChannel(retention=30, cleanfreq=10)
    pro1 = Producer(instid=1, nmsgs=25, msgq=mqueue, rwindow=[2, 3])
    pro2 = Producer(instid=2, nmsgs=15, msgq=mqueue, rwindow=[1, 5])
    pro3 = Producer(instid=3, nmsgs=5, msgq=mqueue, rwindow=[1, 2])
    con1 = Consumer(instid=1, msgq=mqueue, rwindow=[1, 4])
    con2 = Consumer(instid=2, msgq=mqueue, rwindow=[2, 3])
    con3 = Consumer(instid=3, msgq=mqueue, rwindow=[2, 4])

    # Shuffle to start in random order
    funcs = []
    funcs.append(pro1.process())
    funcs.append(pro2.process())
    funcs.append(con1.process())
    funcs.append(con2.process())
    funcs.append(mqueue.cleanqueue())
    random.shuffle(funcs)

    # Create asyncio tasks for execution
    tasks = []
    producer_tasks = []
    for func in funcs:
        await asyncio.sleep(random.randint(1, 5))
        if "Producer" in func.__qualname__:
            producer_tasks.append(loop.create_task(func))
        else:
            tasks.append(loop.create_task(func))

    # Start later consumer and producer respectively
    await asyncio.sleep(35)
    tasks.append(loop.create_task(con3.process()))
    await asyncio.sleep(10)
    producer_tasks.append(loop.create_task(pro3.process()))

    # Wait for all producers to complete
    await asyncio.gather(*producer_tasks)

    # Cancel consumers and message queue clean task
    print("Producers completed, give few minutes for consumers")
    await asyncio.sleep(120)
    for task in tasks:
        task.cancel()
    try:
        await asyncio.gather(*tasks)
    except asyncio.CancelledError:
        print("Consumer cancelled now")

def main():
    '''Main entry function'''
    loop = asyncio.get_event_loop()
    loop.run_until_complete(aio_sessions(loop))

if __name__ == '__main__':
    main()
