"""Using asyncio to simulate message queue"""

import os
import random
import asyncio
from abc import ABC, abstractmethod
from datetime import datetime

class Producer(ABC):
    '''Producer/Publisher base class'''

    @abstractmethod
    async def produce(self):
        '''produce abstract method for producing data'''
        pass

class InMemProducer(Producer):
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

    async def produce(self):
        '''Process function doing main work of producing messages'''
        procid = os.getpid()
        tstamp = datetime.utcnow().strftime("%a, %d %b %Y %H:%M:%S.%f")
        print(f'{tstamp}: Pid[{procid}] {self._pname} begins...')
        low, high = self._rw

        for _ in range(self._num_msgs):
            # Do processing and create message
            await asyncio.sleep(random.randint(low, high))

            msgid, msg = self.create_msg()

            # Put message in the queue
            self.put_msg(msgid=msgid, msg=msg)

        tstamp = datetime.utcnow().strftime("%a, %d %b %Y %H:%M:%S.%f")
        print(f'{tstamp}: Pid[{procid}] {self._pname} done, generated {self._num_msgs} msgs')
